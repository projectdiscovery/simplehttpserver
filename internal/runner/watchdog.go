package runner

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type WatchEvent func(fname string) error

func watchFile(fname string, callback WatchEvent) (watcher *fsnotify.Watcher, err error) {
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := callback(fname); err != nil {
						log.Println("err", err)
					}
				}
			case <-watcher.Errors:
				// ignore errors for now
			}
		}
	}()

	err = watcher.Add(fname)
	return
}
