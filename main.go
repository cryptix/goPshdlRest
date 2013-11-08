package main

import (
	"fmt"
	"os"
	"strings"
)

var usageString = `wrong usage. available commands:
%s open <wid> 	# opens an exisitng workspace and downloads its pshdl code
%s new <dir>		# creates a new workspace from the specified directory
%s stream <wid>	# streams events from an exisitng workspace
`

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, usageString, os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {

	case "open":
		wp, err := OpenWorkspace(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("WP Open:", wp)

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
		fmt.Println("WP Open:", wp)

		done := make(chan bool, 1)
		err = wp.OpenEventStream(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Iterating over events...")
		go func() {
			for {
				select {
				case ev := <-wp.Events:
					subj := ev.GetSubject()
					switch {

					case strings.HasPrefix(subj, "P:WORKSPACE:"):
						fmt.Println("Worskpace Changed:", subj)
						for _, file := range ev.GetFiles() {
							fmt.Printf("[*] %s\n", file.RelPath)
						}

					case subj == "P:COMPILER:VHDL":
						fmt.Println("New VHDL Code")
						errc := make(chan error)
						files := ev.GetFiles()
						count := len(files)

						if count == 0 {
							continue
						}

						for _, file := range files {
							fmt.Printf("[*] Downloading %s\n", file.RelPath)
							// ugly...
							go func(f PshdlApiRecord) {
								f.DownloadFile(errc)
							}(file)
						}

						for err := range errc {
							if err != nil {
								fmt.Fprintf(os.Stderr, "Could not load all files. %s", err)
								break
							} else {
								count -= 1
								if count == 0 {
									fmt.Println("[*] Download finished..")
									close(errc)
								}
							}
						}
					case subj == "P:COMPILER:C":
						fmt.Println("New C-Sim Code")
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
