package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/cryptix/goPshdlRest/api"
	"github.com/jaschaephraim/lrserver"
	"github.com/skratchdot/open-golang/open"
)

var (
	apiClient *pshdlApi.Client
	workspace *pshdlApi.Workspace
)

func main() {
	app := cli.NewApp()
	app.Name = "pshdlViewer"
	app.Usage = "visualize a workspace with it's modules and ports"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "workspace,w", Value: "", Usage: "The workspace to open"},
		cli.StringFlag{Name: "host", Value: "localhost", Usage: "The http host to listen on"},
		cli.IntFlag{Name: "port,p", Value: 3003, Usage: "The http port to listen on"},
	}
	app.Action = run

	app.Run(os.Args)

}

func run(c *cli.Context) {

	wid := c.String("workspace")
	if wid == "" {
		log.Println("please supply a workspace id")
		os.Exit(1)
	}

	// Start LiveReload server
	go lrserver.ListenAndServe()

	apiClient = pshdlApi.NewClientWithID(nil, wid)

	updateWorkspace()

	evChan, err := apiClient.Streaming.OpenEventStream()
	check(err)
	log.Println("EventStream open")

	// Start goroutine that requests reload upon watcher event
	go func() {
		tick := time.Tick(time.Minute * 1)
		for {
			select {
			case ev := <-evChan:
				subj := ev.GetSubject()
				log.Println("[R]", subj)

				switch {
				case strings.HasPrefix(subj, "P:WORKSPACE:"):
					for _, file := range ev.GetFiles() {
						log.Printf("[*] %s\n", file.RelPath)
					}
					go updateWorkspace()

				}
			case <-tick:
				go updateWorkspace()
			}
		}

	}()

	// Start serving html
	http.HandleFunc("/", indexHandler)
	// http.HandleFunc("/md", mdHandler)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))

	listenAddr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))

	done := make(chan struct{})
	go func() {
		err := http.ListenAndServe(listenAddr, nil)
		check(err)
		close(done)
	}()

	err = open.Run("http://" + listenAddr)
	check(err)

	<-done
}

func updateWorkspace() {
	var err error
	start := time.Now()
	workspace, _, err = apiClient.Workspace.GetInfo()
	check(err)
	log.Println("GetInfo() returned after:", time.Since(start))
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
