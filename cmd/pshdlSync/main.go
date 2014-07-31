package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/cryptix/goPshdlRest/api"
	"github.com/visionmedia/go-debug"
	"gopkg.in/fsnotify.v0"
)

const (
	appName  = "pshdlSync"
	widFname = ".wid"
)

var dbg = debug.Debug(appName)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "sync a remote PSHDL workspace with your filesystem"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "workspace,w", Usage: "specifiy the workspace to connect to"},
	}
	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) {
	var (
		err    error
		client *pshdlApi.Client
		wp     *pshdlApi.Workspace
	)

	widStat, widStatErr := os.Stat(widFname)

	if os.IsNotExist(widStatErr) {
		// TODO: pshdlApi.CreateWorkspace()
		wp, _, err = client.Workspace.Create()
		check(err)
		log.Println("Workspace Created:", wp.ID)

		err = ioutil.WriteFile(widFname, []byte(wp.ID), os.ModePerm-7)
		check(err)
	}

	if widStat.Mode().IsRegular() {
		wid, err := ioutil.ReadFile(widFname)
		check(err)

		// TODO: pshdlApi.OpenWorkspace()
		client = pshdlApi.NewClientWithID(nil, string(wid[:16]))
	}

	wp, _, err = client.Workspace.GetInfo()
	check(err)
	log.Printf("Workspace Opened:%s - PID:%d", wp.ID, os.Getpid())
	log.Println("Files:")
	recs := make([]pshdlApi.Record, len(wp.Files))
	for i, f := range wp.Files {
		log.Printf("* %s\nInfos:%s\n", f.Record.RelPath, f.ModuleInfos)
		recs[i] = f.Record
	}

	// todo check if files allready there
	err = client.Workspace.DownloadRecords(recs)
	check(err)
	log.Println("Download of PSHDL-Code complete.")

	//todo push containing files
	log.Println("Starting to watch..")

	watcher, err := fsnotify.NewWatcher()
	check(err)
	defer watcher.Close()

	done := make(chan struct{})

	// Process events
	go func() {
		for err := range watcher.Errors {
			check(err)
			close(done)
		}
	}()

	go func() {
		for ev := range watcher.Events {
			dbg("watcher event: %s", ev)

			if strings.HasSuffix(ev.Name, ".pshdl") {
				switch {
				case ev.Op&fsnotify.Create == fsnotify.Create:
					fallthrough
				case ev.Op&fsnotify.Write == fsnotify.Write:
					log.Println("write to ", ev.Name, ", uploading...")
					file, err := os.Open(ev.Name)
					check(err)

					err = client.Workspace.UploadFile(filepath.Base(ev.Name), file)
					if err != nil {
						log.Fatalf("UploadFile error: %s\n", err)
						return
					}
					file.Close()

					_, err = client.Compiler.Validate()
					if err != nil {
						log.Printf("Validate error: %s\n", err)
					}

				case ev.Op&fsnotify.Remove == fsnotify.Remove:
					log.Println(ev.Name, "deleted, skipping...")
					// client.Workspace.Delete(ev.Name)
				}
			}
		}
		close(done)
	}()

	cwd, err := os.Getwd()
	check(err)

	err = watcher.Add(cwd)
	check(err)

	<-done

	os.Exit(0)
}

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("Fatal from <%s:%d>\n", file, line)
		log.Fatal("Error:", err)
	}
}
