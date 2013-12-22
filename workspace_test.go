package goPshdlRest

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestWorkspaceService_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v0.1/workspace", func(w http.ResponseWriter, r *http.Request) {
		testFormValues(t, r, values{
			"name":  defaultName,
			"eMail": defaultEmail,
		})
		testMethod(t, r, "POST")

		fmt.Fprint(w, `/api/v0.1/workspace/251C5321A7254D79`)
	})

	workspace, _, err := client.Workspace.Create()
	if err != nil {
		t.Errorf("WorkspaceService.Create returned error: %v", err)
	}

	want := &Workspace{Id: "251C5321A7254D79"}
	if !reflect.DeepEqual(workspace, want) {
		t.Errorf("WorkspaceService.Create returned %+v, want %+v", workspace, want)
	}
}

func TestWorkspaceService_GetInfo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v0.1/workspace/1234", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		fmt.Fprint(w, `{"files":[ ], "id":"1234", "lastValidation":0, "jsonVersion":"1.0", "validated":true}`)
	})

	workspace, _, err := client.Workspace.GetInfo("1234")
	if err != nil {
		t.Errorf("Workspace.GetInfo returned error: %v", err)
	}

	want := Workspace{
		Id:          "1234",
		JsonVersion: "1.0",
		Validated:   true,
	}
	if !reflect.DeepEqual(workspace, want) {
		t.Errorf("Workspace.GetInfo returned %+v, want %+v", workspace, want)
	}
}

func TestWorkspaceService_GetInfo_Files(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v0.1/workspace/1234", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		fmt.Fprint(w, `{"files":[ {
    "record" : {
      "fileURI" : "/api/v0.1/workspace/1234/test.pshdl",
      "relPath" : "test.pshdl",
      "lastModified" : 1387740467000
    },
    "syntax" : "unknown",
    "type" : "pshdl",
    "moduleInfos" : [ ]
  }], "id":"1234", "lastValidation":0, "jsonVersion":"1.0", "validated":true}`)
	})

	workspace, _, err := client.Workspace.GetInfo("1234")
	if err != nil {
		t.Errorf("Workspace.GetInfo returned error: %v", err)
	}

	want := []File{
		File{
			Syntax:      "unknown",
			Type:        "pshdl",
			ModuleInfos: []ModuleInfos{},
			Record: Record{
				FileURI:      "/api/v0.1/workspace/1234/test.pshdl",
				RelPath:      "test.pshdl",
				LastModified: 1387740467000,
			},
		},
	}
	if !reflect.DeepEqual(workspace.Files, want) {
		t.Errorf("Workspace.GetInfo returned %+v, want %+v", workspace.Files, want)
	}
}
