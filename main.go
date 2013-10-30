package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/fsnotify"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <open|new> <id|path>\n", os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {
	case "open":
		wp, err := OpenWorkspace(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("WP Found:", wp)
		fmt.Println("Files:", wp.Files)
		// copy files to temp dir
	case "new":
		wp, err := NewWorkspace("Henry", "crypt@me.com")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("WP Created:", wp.Id)
		watchWorskpace(os.Args[2], wp)
	}

}

func watchWorskpace(dir string, wp *PshdlWorkspace) {
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
					err := wp.AddFile(ev.Name)
					if err != nil {
						panic(err)
					}
				}
				// fmt.Println("event:", ev)
			case err := <-watcher.Error:
				fmt.Println("error:", err)
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

	/* ... do stuff ... */
	watcher.Close()
}
