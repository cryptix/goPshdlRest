package main

import (
	"log"
	"os"
	"path/filepath"
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
		cli.StringFlag{Name: "lang,l", Value: "c", Usage: "In which language"},
		cli.StringFlag{Name: "workspace,w", Value: "", Usage: "The workspace to use"},
		cli.StringFlag{Name: "module,m", Value: "", Usage: "The module that should be requested"},
		cli.BoolFlag{Name: "base,b", Usage: "strip the relPath to it's base"},
		cli.StringFlag{Name: "dir,d", Value: "", Usage: "The dir where downloaded code should be stored"},
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

	// TODO push this into the api
	var simLang pshdlApi.SimCodeType
	switch c.String("lang") {
	case "c":
		simLang = pshdlApi.SimC
	case "go":
		simLang = pshdlApi.SimGo
	default:
		log.Println("Unknown SimCodeType")
		os.Exit(1)
	}

	uris, err := apiClient.Compiler.RequestSimCode(simLang, moduleName)
	check(err)

	// construct []Record
	recs := make([]pshdlApi.Record, len(uris))
	for i, uri := range uris {
		uriParts := strings.Split(uri, "/")
		relPath := strings.Replace(uriParts[len(uriParts)-1], ":", "/", -1)
		if c.Bool("base") {
			relPath = filepath.Base(relPath)
		}

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
