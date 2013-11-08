package main

import (
	"fmt"
	"os"
	"strings"
)

var cmdStream = &Command{
	Run:       runStream,
	UsageLine: "stream [flags] workspaceId",
	Short:     "hooks into server-sent events for changes to the workspace",
	Long: `
This automatically fetches changes to the workspace. Like new generated VHDL or Simulation code.
`,
}

func init() {
	fmt.Println("TODO: Set up <stream> flags")
}

func runStream(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
	}

	wp, err := OpenWorkspace(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(1)
		return
	}
	fmt.Println("WP Open:", wp)

	done := make(chan bool, 1)
	err = wp.OpenEventStream(done)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(1)
		return
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
					// errc := make(chan error)
					// files := ev.GetFiles()
					// count := len(files)

					// if count == 0 {
					// 	continue
					// }

					// for _, file := range files {
					// 	fmt.Printf("[*] Downloading %s\n", file.RelPath)
					// 	// ugly...
					// 	go func(f PshdlApiRecord) {
					// 		f.DownloadFile(errc)
					// 	}(file)
					// }

					// for err := range errc {
					// 	if err != nil {
					// 		fmt.Fprintf(os.Stderr, "Could not load all files. %s", err)
					// 		break
					// 	} else {
					// 		count -= 1
					// 		if count == 0 {
					// 			fmt.Println("[*] Download finished..")
					// 			close(errc)
					// 		}
					// 	}
					// }
				case subj == "P:COMPILER:C":
					fmt.Println("New C-Sim Code")
				}
			}
		}
	}()

	<-done
}
