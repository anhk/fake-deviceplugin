package main

import (
	"context"
	dp "fake-deviceplugin/app/deviceplugin"
	"fake-deviceplugin/app/scheduler"
	"fake-deviceplugin/app/volcano"
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func WaitKubeletRestart() {
	ctx := utils.GetInitContext()
	watcher, err := fsnotify.NewWatcher()
	utils.PanicIfError(err)
	defer watcher.Close()

	log.Debug(ctx, "wait for notify kubelet.sock")

	utils.PanicIfError(watcher.Add(pluginapi.KubeletSocket))
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Infof(ctx, "inotify: %s removed, restarting.", pluginapi.KubeletSocket)
				os.Exit(255)
			}
		case err := <-watcher.Errors:
			log.Infof(ctx, "inotify error: %v", err)
		}
	}
}

func StartVolcanoAnnotationSync(ctx context.Context, nodeName string, gpus []volcano.GPUInfo, interval time.Duration) {
	utils.PanicIfError(volcano.SyncNodeAnnotations(ctx, nodeName, gpus))
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := volcano.SyncNodeAnnotations(ctx, nodeName, gpus); err != nil {
				log.Errorf(ctx, "periodic sync volcano annotations failed: %v", err)
			}
		}
	}()
}

func main() {
	// 支持设备插件注册到volcano的节点级别注解
	ctx := utils.GetInitContext()
	nodeName := volcano.ResolveNodeName()
	gpus := volcano.BuildFakeGPUs(16, 10, 16380, "NVIDIA-NVIDIA GeForce RTX 4060 Ti")
	StartVolcanoAnnotationSync(ctx, nodeName, gpus, 30*time.Second)

	// 设备插件注册资源
	dp := dp.NewDevicePlugin("xxfe.com/fake-device", 16)
	dp.Start()

	// 调度扩展接口
	sched := scheduler.NewScheduler()
	sched.Start()

	// 如果kubelet重启了,设备插件需要重新注册,因此监听kubelet.sock的删除事件,一旦检测到就退出,让容器重启器来重启设备插件
	WaitKubeletRestart()
}
