package main

import (
	"html/template"
	"net/http"
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

// indexHandler builds a list with links to all .md files in the watchDir
func indexHandler(rw http.ResponseWriter, req *http.Request) {

	rw.WriteHeader(http.StatusOK)
	if err := tmplFiles.Execute(rw, workspace); err != nil {
		panic(err)
	}
}
