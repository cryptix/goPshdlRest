package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cryptix/goPshdlRest/api"
)

var (
	streamVHDL = flag.Bool("vhdl", false, "download generated vhdl")
	streamCSim = flag.Bool("csim", false, "download generated C Simulation code")
)

func main() {
	flag.Parse()

	if flag.NArg() != 0 {
		log.Fatal(`
		This automatically fetches changes to the workspace.

		Use these flags to download the wanted files.
		-vhdl 	For generated VHDL
		-csim 	For generated C Simulation
		`)
	}

	if *streamCSim {
		log.Println("Csim download active")
	}

	if *streamVHDL {
		log.Println("Csim download active")
	}

	// check if we have a workspace id file
	widFile, err := os.Open(".wid")
	if os.IsNotExist(err) {
		log.Println("No <.wid> file. Exiting")
		os.Exit(0)
	} else if err != nil {
		if err != nil {
			log.Fatalf("os.Open(.wid) Error: %s\n", err)
		}
	}

	wid, err := ioutil.ReadAll(widFile)
	widFile.Close()

	client := pshdlApi.NewClientWithID(nil, string(wid))

	evChan, err := client.Streaming.OpenEventStream()
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}

	log.Println("Displaying events...")
	for ev := range evChan {
		subj := ev.GetSubject()
		switch {

		case strings.HasPrefix(subj, "P:WORKSPACE:"):
			log.Println("Worskpace Changed:", subj)
			for _, file := range ev.GetFiles() {
				log.Printf("[*] %s\n", file.RelPath)
			}

		case subj == "P:COMPILER:VHDL" && *streamVHDL:
			log.Println("New VHDL Code")

			err = client.Workspace.DownloadRecords(ev.GetFiles())
			if err != nil {
				log.Fatalf("Workspace.DownloadRecords() Error:. %s", err)
				break
			}
			log.Println("[*] Download finished..")

		case subj == "P:COMPILER:C" && *streamCSim:
			log.Println("New C-Sim Code")
			err = client.Workspace.DownloadRecords(ev.GetFiles())
			if err != nil {
				log.Fatalf("Could not load all files. %s", err)
				break
			}
			log.Println("[*] Download finished..")
		}
	}
}
