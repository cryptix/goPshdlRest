package goPshdlRest

import (
	"bytes"
	"fmt"
	"io"
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

		fmt.Fprint(w, "/api/v0.1/workspace/251C5321A7254D79")
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

		fmt.Fprint(w, `{"files":[ ], "id":"1234", "lastValidation":10, "jsonVersion":"9.9", "validated":true}`)
	})

	workspace, _, err := client.Workspace.GetInfo()
	if err != nil {
		t.Errorf("Workspace.GetInfo returned error: %v", err)
	}

	want := &Workspace{
		Id:             "1234",
		JsonVersion:    "9.9",
		LastValidation: 10,
		Validated:      true,
		Files:          []File{},
	}
	if !reflect.DeepEqual(workspace, want) {
		t.Errorf("Workspace.GetInfo returned %#v, want %#v", workspace, want)
	}
}

func TestWorkspaceService_GetInfo_Files(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v0.1/workspace/1234", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		fmt.Fprint(w, `{"files":[ { "record" : { "fileURI" : "/api/v0.1/workspace/1234/test.pshdl", "relPath" : "test.pshdl","lastModified" : 1387740467000}, "syntax" : "unknown","type" : "pshdl","moduleInfos" : [ ] }], "id":"1234", "lastValidation":0, "jsonVersion":"1.0", "validated":true}`)
	})

	workspace, _, err := client.Workspace.GetInfo()
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

func TestWorkspaceService_Delete(t *testing.T) {
	setup()
	defer teardown()

	id := "1234"
	fname := "test.pshdl"

	testUrl := fmt.Sprintf("/api/v0.1/workspace/%s/%s", id, fname)
	mux.HandleFunc(testUrl, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		http.Error(w, "", http.StatusOK)
	})

	done, _, err := client.Workspace.Delete(id, fname)
	if err != nil {
		t.Errorf("Workspace.Delete returned error: %v", err)
	}

	if done == false {
		t.Errorf("Worksace.Delete did not set done correctly")
	}
}

func TestWorkspaceService_UploadFile(t *testing.T) {
	setup()
	defer teardown()

	id := "1234"
	fname := "test.pshdl"
	content := []byte("module test {}")

	testUrl := fmt.Sprintf("/api/v0.1/workspace/%s/%s", id, fname)
	mux.HandleFunc(testUrl, func(w http.ResponseWriter, r *http.Request) {
		testHeader(t, r, "Accept", "text/plain")
		testMethod(t, r, "POST")

		upload, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("Workspace.UploadFile request.FormFile error:", err)
		}

		var buf bytes.Buffer

		io.Copy(&buf, upload)
		if buf.String() != string(content) {
			t.Errorf("Workspace.UploadFile upload incompleted\nGot %+v, Want: %+v\n", buf.String(), string(content))
		}

		http.Error(w, "", http.StatusOK)
	})

	inputBuf := bytes.NewReader(content)
	err := client.Workspace.UploadFile(id, fname, inputBuf)
	if err != nil {
		t.Errorf("Workspace.UploadFile returned error: %v", err)
	}
}

func TestWorkspaceService_DownloadFile(t *testing.T) {
	setup()
	defer teardown()

	id := "1234"
	fName := "test.pshdl"
	fContent := "module test {}"

	testUrl := fmt.Sprintf("/api/v0.1/workspace/%s/%s", id, fName)
	mux.HandleFunc(testUrl, func(w http.ResponseWriter, r *http.Request) {
		testHeader(t, r, "Accept", "text/plain")
		testMethod(t, r, "GET")

		fmt.Fprintf(w, fContent)
	})

	fResponse, err := client.Workspace.DownloadFile(id, fName)
	if err != nil {
		t.Errorf("Workspace.DownloadFile returned error: %v", err)
	}

	if string(fResponse) != fContent {
		t.Errorf("Workspace.DownloadFile incorrect file Download.\nResponse:%s\nWanted:%s\n", string(fResponse), fContent)
	}
}
