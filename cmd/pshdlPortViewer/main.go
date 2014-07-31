package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/carbocation/interpose"
	"github.com/codegangsta/cli"
	"github.com/cryptix/goPshdlRest/api"
	"github.com/gorilla/mux"
	"github.com/jaschaephraim/lrserver"
	"github.com/skratchdot/open-golang/open"
	"github.com/visionmedia/go-debug"
)

var (
	apiClient *pshdlApi.Client
	workspace *pshdlApi.Workspace
)

var dbg = debug.Debug("pshdlPortViewer")

func main() {
	app := cli.NewApp()
	app.Name = "pshdlPortViewer"
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
	lrs, err := lrserver.NewLRServer(nil)
	check(err)

	apiClient = pshdlApi.NewClientWithID(nil, wid)

	updateWorkspace()

	evChan, err := apiClient.Streaming.OpenEventStream()
	check(err)
	log.Println("EventStream open")

	// Start goroutine that requests reload upon watcher event
	go func() {
		for ev := range evChan {
			subj := ev.GetSubject()
			log.Println("workspace event:", subj)

			switch {
			case strings.HasPrefix(subj, "P:WORKSPACE:"):
				updateWorkspace()
				lrs.Reload("workspaceUpdate")
			}
		}
	}()

	listenAddr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))

	// Start serving html
	errc := make(chan error)
	go func() {
		select {
		case e := <-errc:
			check(e)
		case <-time.After(time.Millisecond * 150):
			log.Println("Opening Browser")
			err = open.Run("http://" + listenAddr)
			check(err)
		}

	}()
	middle := interpose.New()

	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	r.HandleFunc("/validate", validateHandler)
	middle.UseHandler(r)

	// Tell the browser which server this came from
	middle.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(rw, req)
			dbg("http: %s (%v)", req.URL.Path, time.Since(start))
		})
	})

	errc <- http.ListenAndServe(listenAddr, middle)
	check(err)
	close(errc)

}

func updateWorkspace() {
	var err error
	start := time.Now()
	workspace, _, err = apiClient.Workspace.GetInfo()
	if err != nil {
		log.Printf("updateWorkspace Failed: %s (%v)\n", err, time.Since(start))
		return
	}
	dbg("updateWorkspace (%v)", time.Since(start))
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
