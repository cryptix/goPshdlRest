package main

import (
	"fmt"
	"os"
)

var cmdOpen = &Command{
	Run:       runOpen,
	UsageLine: "open [flags] workspaceId",
	Short:     "download and watch an existing workspace",
	Long: `
Downloads .pshdl files from the workspace to disk and starts watching them for changes.
	`,
}

func init() {
	fmt.Println("TODO: Set up <open> flags")
}

func runOpen(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
	}

	wp, err := OpenWorkspace(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(1)
		return
	}
	fmt.Println("WP Opened:", wp)

	fmt.Println("Files:", wp.Files)
	err = wp.DownloadAllFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(2)
		return
	}

	fmt.Println("All files Downloaded, watching..")
	wp.watch(wp.Id)
}
