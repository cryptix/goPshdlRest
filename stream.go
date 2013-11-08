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
This automatically fetches changes to the workspace.

Use these flags to download the wanted files.
-vhdl 	For generated VHDL

-csim 	For generated C Simulation
`,
}

var (
	streamVHDL bool
	streamCSim bool
)

func init() {
	cmdStream.Flag.BoolVar(&streamVHDL, "vhdl", false, "download generated vhdl")
	cmdStream.Flag.BoolVar(&streamCSim, "csim", false, "download generated C Simulation code")
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

	if streamVHDL {
		fmt.Println("Downloading generated VHDL...")
		go func() {
			for ev := range wp.Events {
				subj := ev.GetSubject()
				if subj == "P:COMPILER:VHDL" {
					err := ev.DownloadFiles()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Could not load all files. %s", err)
						break
					}
					fmt.Println("[*] Download finished..")
				}
			}
		}()
	}

	if streamCSim {
		fmt.Println("Downloading generated C Simulation code...")
		go func() {
			for ev := range wp.Events {
				subj := ev.GetSubject()
				if subj == "P:COMPILER:C" {
					err := ev.DownloadFiles()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Could not load all files. %s", err)
						break
					}
					fmt.Println("[*] Download finished..")
				}
			}
		}()
	}

	fmt.Println("Displaying events...")
	go func() {
		for ev := range wp.Events {
			subj := ev.GetSubject()
			switch {

			case strings.HasPrefix(subj, "P:WORKSPACE:"):
				fmt.Println("Worskpace Changed:", subj)
				for _, file := range ev.GetFiles() {
					fmt.Printf("[*] %s\n", file.RelPath)
				}

			case subj == "P:COMPILER:VHDL":
				fmt.Println("New VHDL Code")

			case subj == "P:COMPILER:C":
				fmt.Println("New C-Sim Code")
			}
		}
	}()

	<-done
}
