package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
)

func watch(dir, id string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if strings.HasSuffix(ev.Name, ".pshdl") {
					fmt.Println("PSHDL Code! Adding", ev.Name)

					file, err := os.Open(ev.Name)
					if err != nil {
						fmt.Fprintf(os.Stderr, "os.Open error: %s\n", err)
						done <- true
						return
					}

					err = client.Workspace.UploadFile(filepath.Base(ev.Name), file)
					if err != nil {
						fmt.Fprintf(os.Stderr, "UploadFile error: %s\n", err)
						done <- true
						return
					}
					file.Close()
				}
			case err := <-watcher.Error:
				fmt.Println("error:", err)
				done <- true
			}
		}
	}()

	err = watcher.Watch(dir)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("Invalid Watch Directory..")
			os.Exit(1)
		}
		panic(err)
	}

	<-done

	watcher.Close()
}
