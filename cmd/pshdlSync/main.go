package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptix/goPshdlRest/api"
	"gopkg.in/fsnotify.v0"
)

func main() {
	var wp *pshdlApi.Workspace
	client := pshdlApi.NewClient(nil)

	// check if we have a workspace id file
	widFile, err := os.Open(".wid")
	if err == nil {
		wid, err := ioutil.ReadAll(widFile)

		client = pshdlApi.NewClient(nil)
		client.Workspace.ID = string(wid[:16])
		client.Compiler.ID = string(wid[:16])

		wp, _, err = client.Workspace.GetInfo()
		if err != nil {
			log.Fatalf("Workspace.GetInfo() API Error: %s\n", err)
		}
		log.Printf("Workspace Opened:%s", wp.ID)

		log.Println("Files:")
		recs := make([]pshdlApi.Record, len(wp.Files))
		for i, f := range wp.Files {
			log.Println("*", f.Record.RelPath)
			recs[i] = f.Record
		}

		// todo check if files allready there
		err = client.Workspace.DownloadRecords(recs)
		if err != nil {
			log.Fatalf("Workspace.DownloadRecords() API Error: %s\n", err)
		}
		log.Println("Download of PSHDL-Code complete.")

	} else if os.IsNotExist(err) {
		log.Println("No <.wid> file, create new workspace")

		wp, _, err = client.Workspace.Create()
		if err != nil {
			log.Fatalf("Workspace.Create() API Error: %s\n", err)
		}
		log.Println("Workspace Created:", wp.ID)

	} else if err != nil {
		if err != nil {
			log.Fatalf("os.Open(.wid) Error: %s\n", err)
		}
	}
	widFile.Close()

	//todo push containing files
	log.Println("Starting to watch..")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan struct{})

	go func() {
		for err := range watcher.Errors {
			close(done)
			log.Fatal(err)
		}
	}()

	// Process events
	go func() {
		for ev := range watcher.Events {

			log.Println("event:", ev)

			if strings.HasSuffix(ev.Name, ".pshdl") {
				switch {

				case ev.Op&fsnotify.Write == fsnotify.Write:
					log.Println("write to ", ev.Name, ", uploading...")
					file, err := os.Open(ev.Name)
					if err != nil {
						log.Fatalf("os.Open error: %s\n", err)
						return
					}

					err = client.Workspace.UploadFile(filepath.Base(ev.Name), file)
					if err != nil {
						log.Fatalf("UploadFile error: %s\n", err)
						return
					}
					file.Close()

				case ev.Op&fsnotify.Remove == fsnotify.Remove:
					log.Println(ev.Name, "deleted, removing from api wp...")
					// client.Workspace.Delete(ev.Name)
				}
			}
		}
		close(done)
	}()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(cwd)
	if err != nil {
		log.Fatal(err)
	}

	<-done

	os.Exit(0)
}
