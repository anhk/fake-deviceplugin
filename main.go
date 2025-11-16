package main

import (
	dp "fake-deviceplugin/app/deviceplugin"
	"fake-deviceplugin/app/scheduler"
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"os"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func WaitKubeletRestart() {
	ctx := utils.GetInitContext()
	watcher, err := fsnotify.NewWatcher()
	utils.PanicIfError(err)
	defer watcher.Close()

	utils.PanicIfError(watcher.Add(pluginapi.KubeletSocket))
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Infof(ctx, "inotify: %s created, restarting.", pluginapi.KubeletSocket)
				os.Exit(255)
			}
		case err := <-watcher.Errors:
			log.Infof(ctx, "inotify error: %v", err)
		}
	}
}

func main() {
	dp := dp.NewDevicePlugin("xxfe.com/fake-device", 4)
	dp.Start()

	sched := scheduler.NewScheduler()
	utils.PanicIfError(sched.Start())

	WaitKubeletRestart()
}
