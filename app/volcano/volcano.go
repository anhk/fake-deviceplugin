// Package volcano 提供设备插件与Volcano调度器的集成功能
// 该模块负责为节点添加GPU相关的注解,使得Volcano调度器能够识别和管理节点上的GPU资源
//
// 主要功能:
// - volcano.sh/node-vgpu-handshake: 节点GPU握手信息,格式为"Requesting_<timestamp>",用于标识节点与设备插件的通信状态
// - volcano.sh/node-vgpu-register: GPU设备注册信息,包含以下字段:
//   - GPU唯一标识符 (如: GPU-ddb54984-de77-c40b-6046-612894b31b08)
//   - GPU核心数量 (如: 10)
//   - GPU显存大小(MB) (如: 16380)
//   - GPU型号名称 (如: NVIDIA-NVIDIA GeForce RTX 4060 Ti)
//   - GPU可用状态 (如: true/false)
//   - GPU分配信息 (如: hami-core:)
//
// 这些注解使Volcano调度器能够准确感知节点的GPU资源,从而进行更有效的调度决策
package volcano

import (
	"context"
	"encoding/json"
	"fake-deviceplugin/pkg/k8s"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	NodeVGPUHandshakeAnnotation = "volcano.sh/node-vgpu-handshake"
	NodeVGPURegisterAnnotation  = "volcano.sh/node-vgpu-register"
)

// GPUInfo 描述一张GPU在Volcano注解中的注册信息.
type GPUInfo struct {
	ID         string
	Core       int
	MemoryMB   int
	Model      string
	Healthy    bool
	Allocation string
}

func newNvidiaLikeGPUUUID() string {
	return "GPU-" + strings.ToUpper(uuid.NewString())
}

func (g GPUInfo) normalize() GPUInfo {
	if g.ID == "" {
		g.ID = newNvidiaLikeGPUUUID()
	}
	if g.Model == "" {
		g.Model = "NVIDIA-Unknown"
	}
	if g.Allocation == "" {
		g.Allocation = "hami-core"
	}
	return g
}

func (g GPUInfo) formatForRegister() string {
	g = g.normalize()
	return fmt.Sprintf("%s,%d,%d,%s,%t,%s", g.ID, g.Core, g.MemoryMB, g.Model, g.Healthy, g.Allocation)
}

func buildHandshake() string {
	return "Requesting_" + time.Now().UTC().Format("2006.01.02 15:04:05")
}

func buildRegisterValue(gpus []GPUInfo) string {
	if len(gpus) == 0 {
		return ""
	}
	parts := make([]string, 0, len(gpus))
	for _, gpu := range gpus {
		parts = append(parts, gpu.formatForRegister())
	}
	return strings.Join(parts, ":")
}

// BuildFakeGPUs 生成指定数量的演示GPU信息，便于快速接入或本地测试.
func BuildFakeGPUs(count, core, memoryMB int, model string) []GPUInfo {
	if count <= 0 {
		return nil
	}
	items := make([]GPUInfo, 0, count)
	for i := 0; i < count; i++ {
		items = append(items, GPUInfo{
			ID:       newNvidiaLikeGPUUUID(),
			Core:     core,
			MemoryMB: memoryMB,
			Model:    model,
			Healthy:  true,
		})
	}
	return items
}

// ResolveNodeName 解析当前节点名，优先读取NODE_NAME，其次读取HOSTNAME.
func ResolveNodeName() string {
	if nodeName := strings.TrimSpace(os.Getenv("NODE_NAME")); nodeName != "" {
		return nodeName
	}
	if hostname := strings.TrimSpace(os.Getenv("HOSTNAME")); hostname != "" {
		return hostname
	}
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(hostname)
}

// SyncNodeAnnotations 将Volcano需要的GPU注解同步到指定节点.
func SyncNodeAnnotations(ctx context.Context, nodeName string, gpus []GPUInfo) error {
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" {
		return fmt.Errorf("node name is empty")
	}

	annotations := map[string]string{
		NodeVGPUHandshakeAnnotation: buildHandshake(),
		NodeVGPURegisterAnnotation:  buildRegisterValue(gpus),
	}

	patchObj := map[string]any{
		"metadata": map[string]any{
			"annotations": annotations,
		},
	}
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return err
	}

	_, err = k8s.GetKubeClient(ctx).CoreV1().Nodes().Patch(ctx, nodeName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	// log.Infof(ctx, "sync volcano annotations ok, node=%s handshake=%s register=%s",
	// 	nodeName, annotations[NodeVGPUHandshakeAnnotation], annotations[NodeVGPURegisterAnnotation])
	return nil
}
