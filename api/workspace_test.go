package pshdlApi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkspaceService(t *testing.T) {

	Convey("Given a clean test server for the WorkspaceService", t, func() {
		setup()

		Convey("Create() should return a new Workspace", func() {

			mux.HandleFunc("/api/v0.1/workspace", func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()

				So(r.Form, ShouldResemble, url.Values{
					"name":  []string{defaultName},
					"eMail": []string{defaultEmail},
				})

				So(r.Method, ShouldEqual, "POST")

				fmt.Fprint(w, "/api/v0.1/workspace/251C5321A7254D79")
			})

			workspace, _, err := client.Workspace.Create()
			So(err, ShouldBeNil)
			So(workspace, ShouldResemble, &Workspace{ID: "251C5321A7254D79"})
		})

		Convey("GetInfo()", func() {

			mux.HandleFunc("/api/v0.1/workspace/1234",
				func(w http.ResponseWriter, r *http.Request) {
					So(r.Method, ShouldEqual, "GET")

					fmt.Fprint(w, `{"files":[ { "record" : { "fileURI" : "/api/v0.1/workspace/1234/test.pshdl", "relPath" : "test.pshdl","lastModified" : 1387740467000}, "syntax" : "unknown","type" : "pshdl","moduleInfos" : [ ] }], "ID":"1234", "lastValIDation":0, "jsonVersion":"1.0", "valIDated":true}`)
				})

			Convey("should return meta workspace info", func() {

				workspace, _, err := client.Workspace.GetInfo()
				So(err, ShouldBeNil)
				So(workspace.ID, ShouldEqual, "1234")
				So(workspace.JSONVersion, ShouldEqual, "1.0")
				So(workspace.LastValIDation, ShouldEqual, 0)
				So(workspace.ValIDated, ShouldBeTrue)
			})

			Convey("should decode the File info", func() {

				workspace, _, err := client.Workspace.GetInfo()
				So(err, ShouldBeNil)
				So(workspace.Files, ShouldResemble, []File{
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
				})
			})
		})

		Convey("Delete()", func() {

			Convey("should return the status of the operation", func() {
				fname := "test.pshdl"
				client.Workspace.ID = "1234"

				mux.HandleFunc(fmt.Sprintf("/api/v0.1/workspace/%s/%s", client.Workspace.ID, fname),
					func(w http.ResponseWriter, r *http.Request) {
						So(r.Method, ShouldEqual, "DELETE")
						// TODO: karsten nerven das leerer output nich klar geht
						http.Error(w, "{}", http.StatusOK)
					})

				done, _, err := client.Workspace.Delete(fname)
				So(err, ShouldBeNil)
				So(done, ShouldBeTrue)
			})

			Convey("without an ID should return an error", func() {
				_, _, err := client.Workspace.Delete("hansfranz.pshdl")
				So(err, ShouldNotBeNil)
			})

		})

		Convey("UploadFile()", func() {

			Convey("with a correct request should return err == nil", func() {

				client.Workspace.ID = "1234"
				fname := "test.pshdl"
				content := []byte("module test {}")

				mux.HandleFunc(fmt.Sprintf("/api/v0.1/workspace/%s", client.Workspace.ID),
					func(w http.ResponseWriter, r *http.Request) {
						So(r.Method, ShouldEqual, "POST")
						So(r.Header.Get("Accept"), ShouldEqual, "text/plain")

						So(r.Header.Get("Content-Type"), ShouldStartWith, "multipart/form-data")

						upload, _, err := r.FormFile("file")
						So(err, ShouldBeNil)

						var buf bytes.Buffer

						_, err = io.Copy(&buf, upload)
						So(err, ShouldBeNil)
						So(buf.String(), ShouldEqual, string(content))

						// TODO: karsten nerven das leerer output nich klar geht
						http.Error(w, "{}", http.StatusOK)
					})

				err := client.Workspace.UploadFile(fname, bytes.NewReader(content))
				So(err, ShouldBeNil)
			})

			Convey("without an ID should return an error", func() {
				err := client.Workspace.UploadFile("hansfranz.pshdl", bytes.NewReader([]byte("")))
				So(err, ShouldNotBeNil)
			})
		})

		Convey("DownloadRecord() with a valID request", func() {
			client.Workspace.ID = "1234"
			fName := "test.pshdl"
			fContent := "module test {}"

			testURL := fmt.Sprintf("/api/v0.1/workspace/%s/%s", client.Workspace.ID, fName)
			mux.HandleFunc(testURL, func(w http.ResponseWriter, r *http.Request) {
				So(r.Header.Get("Accept"), ShouldEqual, "text/plain")
				So(r.Method, ShouldEqual, "GET")

				fmt.Fprintf(w, fContent)
			})

			err := client.Workspace.DownloadRecord(Record{FileURI: testURL, RelPath: fName})
			So(err, ShouldBeNil)
		})

		Convey("DownloadRecord() without an ID", func() {
			err := client.Workspace.DownloadRecord(Record{RelPath: "hansfranz.pshdl"})
			So(err, ShouldNotBeNil)
		})

		Reset(teardown)
	})
}
