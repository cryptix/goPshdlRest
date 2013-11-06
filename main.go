package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/fsnotify"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <open|new|stream> <id|path>\n", os.Args[0])
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
		err = wp.DownloadAllFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("All files Downloaded, watching..")
		wp.watch(wp.Id)

	case "new":
		wp, err := NewWorkspace("JohnGo", "none@me.com")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("WP Created:", wp.Id)
		wp.watch(os.Args[2])

	case "stream":
		wp, err := OpenWorkspace(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		done := make(chan bool)
		err = wp.OpenEventStream(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Iterating over events")
		go func() {
			for {
				select {
				case ev := <-wp.Events:

					switch ev.GetSubject() {

					case "P:WORKSPACE:UPDATED":
						fmt.Println("Worskpace Updated")
						for _, file := range ev.GetFiles() {
							fmt.Printf("[*] %s\n", file.RelPath)
						}

					case "P:COMPILER:VHDL":
						fmt.Println("New VHDL")
						for _, file := range ev.GetFiles() {
							fmt.Printf("[*] %s\n", file.RelPath)
						}

					default:
						fmt.Printf("Unhandled Event.\nSubject: %s\n%#v\n", ev.GetSubject(), ev)
					}
				}
			}
		}()

		<-done

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command %s\n", os.Args[1])
		os.Exit(1)

	}
}

func (wp *PshdlWorkspace) watch(dir string) {
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
