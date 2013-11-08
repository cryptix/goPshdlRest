package main

import (
	"fmt"
	"os"
)

var cmdNew = &Command{
	Run:       runNew,
	UsageLine: "new [flags] path",
	Short:     "create a new workspace and watch for changes",
	Long: `
Asks the API for a new workspace.
Monitors <path> for changes.
Uploads changed .pshdl files to the new Workspace.
`,
}

func init() {
	fmt.Println("TODO: Set up <new> flags")
}

func runNew(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
	}

	wp, err := NewWorkspace("JohnGo", "none@me.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(1)
		return
	}
	fmt.Println("WP Created:", wp)

	//todo push containing files
	fmt.Println("Watching..")
	wp.watch(args[0])
}
