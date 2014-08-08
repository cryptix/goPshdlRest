package main

import (
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/cryptix/goPshdlRest/api"

	"github.com/visionmedia/go-debug"
)

var (
	apiClient *pshdlApi.Client
	workspace *pshdlApi.Workspace
)

const appName = "pshdlFetchSimCode"

var dbg = debug.Debug(appName)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "request simulation code and download it"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "workspace,w", Value: "", Usage: "The workspace to use"},
		cli.StringFlag{Name: "module,m", Value: "", Usage: "The module that should be requested"},
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

	moduleName := c.String("module")
	if moduleName == "" {
		log.Println("please supply a module name")
		os.Exit(1)
	}

	apiClient = pshdlApi.NewClientWithID(nil, wid)
	uris, err := apiClient.Compiler.RequestSimCode(pshdlApi.SimC, moduleName)
	check(err)

	// construct []Record
	recs := make([]pshdlApi.Record, len(uris))
	for i, uri := range uris {
		uriParts := strings.Split(uri, "/")
		relPath := strings.Replace(uriParts[len(uriParts)-1], ":", "/", -1)

		recs[i] = pshdlApi.Record{
			FileURI: uri,
			RelPath: relPath,
		}
	}

	check(apiClient.Workspace.DownloadRecords(recs))
	log.Println("Fetched all files")

}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
