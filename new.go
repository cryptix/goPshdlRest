package main

import (
	"fmt"
	"os"

	api "github.com/cryptix/goPshdlRest"
	"github.com/howeyc/fsnotify"
)

var (
	client *api.Client
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

	client = api.NewClient(nil)

	wp, resp, err := client.Workspace.Create()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		setExitStatus(1)
		return
	}
	fmt.Println("WP Created:", wp.Id)

	//todo push containing files
	fmt.Println("Watching..")
	watch(args[0], wp.Id)
}

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

					file, err := os.Open(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "os.Open error: %s\n", err)
						done <- true
						return
					}
					defer file.Close()

					err = client.Workspace.UploadFile(id, ev.Name, file)
					if err != nil {
						fmt.Fprintf(os.Stderr, "UploadFile error: %s\n", err)
						done <- true
						return
					}
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
