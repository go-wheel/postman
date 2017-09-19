package backendhttp

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/lets-go-go/logger"
)

var (
	chanStopWatcher chan bool
)

func init() {
	chanStopWatcher = make(chan bool, 1)
}

// StopTaskWatcher Stop Task Watcher
func StopTaskWatcher() {
	chanStopWatcher <- true
}

// StartTaskWatcher task watcher
func StartTaskWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				logger.Debugf("new watcher event:%v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					// log.Println("modified file:", event.Name)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					// log.Println("create file:", event.Name)

					AppendNewTask(event.Name)
				}
			case err := <-watcher.Errors:
				logger.Errorf("watcher error:%v", err)

			case <-chanStopWatcher:
				return
			}
		}
	}()

	err = watcher.Add(taskPath)
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(rangeUploadTaskPath)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
