package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	api "github.com/cryptix/goPshdlRest"
	"github.com/howeyc/fsnotify"
)

func main() {
	var wp *api.Workspace
	client := api.NewClient(nil)

	// check if we have a workspace id file
	widFile, err := os.Open(".wid")
	if err == nil {
		wid, err := ioutil.ReadAll(widFile)

		client = api.NewClient(nil)
		client.Workspace.ID = string(wid)
		client.Compiler.ID = string(wid)

		wp, _, err = client.Workspace.GetInfo()
		if err != nil {
			log.Fatalf("Workspace.GetInfo() API Error: %s\n", err)
		}
		log.Printf("Workspace Opened:%s\nFiles:%v\n", wp.ID, wp.Files)

		// todo check if files allready there
		err = client.Workspace.DownloadAllFiles()
		if err != nil {
			log.Fatalf("Workspace.DownloadAllFiles() API Error: %s\n", err)
		}
		log.Printf("Download of PSHDL-Code complete")

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
		panic(err)
	}
	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if strings.HasSuffix(ev.Name, ".pshdl") {
					log.Println("PSHDL Code! Adding", ev.Name)

					file, err := os.Open(ev.Name)
					if err != nil {
						log.Fatalf("os.Open error: %s\n", err)
						done <- true
						return
					}

					err = client.Workspace.UploadFile(filepath.Base(ev.Name), file)
					if err != nil {
						log.Fatalf("UploadFile error: %s\n", err)
						done <- true
						return
					}
					file.Close()
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
				done <- true
			}
		}
	}()

	err = watcher.Watch(".")
	if err != nil {
		if os.IsExist(err) {
			log.Fatal("Invalid Watch Directory..")
		}
		panic(err)
	}

	<-done

	watcher.Close()

	os.Exit(0)
}
