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

type DevicePlugin struct {
	pluginapi.UnimplementedDevicePluginServer
	resourceName string
	count        int
}

func NewDevicePlugin(resource string, count int) *DevicePlugin {
	return &DevicePlugin{resourceName: resource, count: count}
}

func (m *DevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	log.Debug("GetDevicePluginOptions called")
	return nil, nil
}

func (m *DevicePlugin) ListAndWatch(_ *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Debug("ListAndWatch called")
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

func (m *DevicePlugin) GetPreferredAllocation(_ context.Context, _ *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	log.Debug("GetPreferredAllocation called")
	return nil, nil
}

func (m *DevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Debugf("Allocate called with request: %s", utils.JsonString(request))
	return &pluginapi.AllocateResponse{ContainerResponses: []*pluginapi.ContainerAllocateResponse{{}}}, nil
}

func (m *DevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	log.Debugf("PreStartContainer called with request: %s", utils.JsonString(request))
	return nil, nil
}

func register(endpoint, resourceName string) {
	conn, err := unixDial(endpoint, 5*time.Second)
	utils.Must(err)
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(sockPath(resourceName)),
		ResourceName: resourceName,
	}
	log.Debugf("register req: %s", utils.JsonString(req))

	_, err = client.Register(context.Background(), req)
	utils.Must(err)
}

func unixDial(endpoint string, timeout time.Duration) (*grpc.ClientConn, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := grpc.DialContext(timeoutCtx, endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.DialTimeout("unix", endpoint, timeout)
		}))
	return c, err
}

// func unixDial(endpoint string, timeout time.Duration) (*grpc.ClientConn, error) {
// 	return grpc.NewClient(endpoint,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
// 			return net.DialTimeout("unix", endpoint, timeout)
// 		}),
// 	)
// }

func (m *DevicePlugin) Start() {
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
	log.Infof("test sock [%s] ok", sockPath(m.resourceName))

	register(pluginapi.KubeletSocket, m.resourceName)
	log.Info("register device plugin ok")
}

func (m *DevicePlugin) Stop() {}

func sockPath(name string) string {
	name = path.Base(name)
	return path.Join(pluginapi.DevicePluginPath, name) + ".sock"
}
