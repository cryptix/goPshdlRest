package main

import (
	"fmt"
	"os"

	api "github.com/cryptix/goPshdlRest"
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

	fi, err := os.Stat(args[0])
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "%s is not a directory.\n", args[0])
		setExitStatus(1)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Stat Error: %s\n", err)
		setExitStatus(1)
		return
	}

	client = api.NewClient(nil)
	client.Workspace.Id = args[0]

	wp, _, err := client.Workspace.GetInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "API Error: %s\n", err)
		setExitStatus(1)
		return
	}
	fmt.Println("WP Opened:", wp.Id)

	fmt.Println("Files:", wp.Files)

	fmt.Println("TODO: DownloadAllFiles")
	// err = wp.DownloadAllFiles()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	// 	setExitStatus(2)
	// 	return
	// }

	fmt.Println("All files Downloaded, watching..")
	watch(args[0], args[0])
	// wp.watch(wp.Id)
}
