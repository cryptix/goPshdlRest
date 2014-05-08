package goPshdlRest

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompilerService(t *testing.T) {
	Convey("Given a clean test server for the CompilerService", t, func() {
		setup()

		Convey("Validate() send the correct request", func() {

			mux.HandleFunc("/api/v0.1/compiler/1234/validate", func(w http.ResponseWriter, r *http.Request) {
				So(r.Method, ShouldEqual, "POST")

				fmt.Fprint(w, "{}")
			})

			err := client.Compiler.Validate()
			So(err, ShouldBeNil)
		})

		Convey("RequestSimCode()", func() {

			Convey("should return an error when moduleName is empty", func() {
				url, err := client.Compiler.RequestSimCode(SimC, "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing moduleName")
				So(url, ShouldEqual, "")
			})

			Convey("should return an error when SimCodeType is unknown", func() {
				url, err := client.Compiler.RequestSimCode(23, "SomeModule")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unsupported SimCodeType:23")
				So(url, ShouldEqual, "")
			})

			Convey("with valid information, should return the dlURL", func() {
				moduleName := "some.module.mname"
				dlURL := "/api/v0.1/workspace/1234/src-gen:psex:c:some.module.mname.c"

				mux.HandleFunc("/api/v0.1/compiler/1234/psex/c", func(w http.ResponseWriter, r *http.Request) {
					So(r.Method, ShouldEqual, "POST")

					r.ParseForm()
					So(r.Form, ShouldResemble, url.Values{
						"module": []string{moduleName},
					})

					http.Error(w, dlURL, http.StatusCreated)
				})

				url, err := client.Compiler.RequestSimCode(SimC, moduleName)
				So(err, ShouldBeNil)
				So(url, ShouldEqual, dlURL) //TODO remove the newline from the returned url
			})

			Convey("with a broken workspace, should return an error and empty url", func() {

				mux.HandleFunc("/api/v0.1/compiler/1234/psex/c", func(w http.ResponseWriter, r *http.Request) {
					So(r.Method, ShouldEqual, "POST")

					http.Error(w, `[{}]`, http.StatusBadRequest)
				})

				url, err := client.Compiler.RequestSimCode(SimC, "abc")
				So(err, ShouldNotBeNil)
				So(url, ShouldEqual, "")
			})
		})

		Reset(teardown)
	})
}
