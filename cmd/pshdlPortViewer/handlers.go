package main

import (
	"html/template"
	"log"
	"net/http"
	"sort"

	"github.com/cryptix/goPshdlRest/api"
)

// tmplFiles template with list of files found in the workspace
var tmplFiles = template.Must(template.New("tmplFiles").Parse(`
<!doctype html>
<html>
<head>
	<title>List of Workspace {{.ID}}</title>
		<script src="http://localhost:35729/livereload.js"></script>
<body>
<h1>Workspace - <small>{{.ID}} <strong>Validated:{{.Validated}}</strong></small> </h1>

{{range .Files}}
			<h3>{{.Record.RelPath}}</h3>
			<pre>{{index .ModuleInfos 0}}</pre>
{{end}}

</body>
</html>
`))

func handler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	if workspace == nil {
		log.Println("workspace out of date..")
		workspace = new(pshdlApi.Workspace)
		return
	}

	for i, file := range workspace.Files {
		ports := file.ModuleInfos[0].Ports
		sort.Sort(pshdlApi.ByDir(ports))
		workspace.Files[i].ModuleInfos[0].Ports = ports
	}

	if err := tmplFiles.Execute(rw, workspace); err != nil {
		panic(err)
	}
}
