package app

import (
	"context"
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type SoftRoceDevicePlugin struct {
	pluginapi.UnimplementedDevicePluginServer
	resourceName string
	count        int
}

func NewSoftRoceDevicePlugin(resource string, count int) *SoftRoceDevicePlugin {
	return &SoftRoceDevicePlugin{resourceName: resource, count: count}
}

func (m *SoftRoceDevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return nil, nil
}

func (m *SoftRoceDevicePlugin) ListAndWatch(_ *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	devices := make([]*pluginapi.Device, 0, m.count)
	for i := 0; i < m.count; i++ {
		devices = append(devices, &pluginapi.Device{
			ID:       fmt.Sprintf("%s-%d", m.resourceName, i),
			Health:   pluginapi.Healthy,
			Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: int64(i)}}},
		})
	}
	utils.Must(server.Send(&pluginapi.ListAndWatchResponse{Devices: devices}))
	select {}
}

func (m *SoftRoceDevicePlugin) GetPreferredAllocation(_ context.Context, _ *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

func (m *SoftRoceDevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	return &pluginapi.AllocateResponse{ContainerResponses: []*pluginapi.ContainerAllocateResponse{{}}}, nil
}

func (m *SoftRoceDevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}

func register(endpoint, resourceName string) {
	conn, err := unixDial(endpoint, 5*time.Second)
	utils.Must(err)
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(resourceName),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), req)
	utils.Must(err)
}

func unixDial(endpoint string, timeout time.Duration) (*grpc.ClientConn, error) {
	return grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.DialTimeout("unix", endpoint, timeout)
		}),
	)
}

func (m *SoftRoceDevicePlugin) Start() {
	utils.Must(os.MkdirAll(pluginapi.DevicePluginPath, 0755))
	_ = unix.Unlink(sockPath(m.resourceName))

	sock, err := net.Listen("unix", sockPath(m.resourceName))
	utils.Must(err)

	server := grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(server, m)

	go func() { utils.Must(server.Serve(sock)) }()
	// Wait for server to start by launching a blocking connection
	conn, err := unixDial(sockPath(m.resourceName), 5*time.Second)
	utils.Must(err)
	utils.Must(conn.Close())
	log.Info("test sock ok")

	register(pluginapi.KubeletSocket, m.resourceName)
	log.Info("register device plugin ok")
}

func (m *SoftRoceDevicePlugin) Stop() {}

func sockPath(name string) string {
	name = path.Base(name)
	return path.Join(pluginapi.DevicePluginPath, name)
}
