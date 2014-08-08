package pshdlApi

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

			_, err := client.Compiler.Validate()
			So(err, ShouldBeNil)
		})

		Convey("RequestSimCode()", func() {

			Convey("should return an error when moduleName is empty", func() {
				uris, err := client.Compiler.RequestSimCode(SimC, "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing moduleName")
				So(uris, ShouldBeNil)
			})

			Convey("should return an error when SimCodeType is unknown", func() {
				uris, err := client.Compiler.RequestSimCode(23, "SomeModule")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unsupported SimCodeType:23")
				So(uris, ShouldBeNil)
			})

			Convey("with valid information, should return the dlURL", func() {
				moduleName := "some.module.mname"
				dlURLs := `/api/v0.1/workspace/1234/src-gen:psex:c:de.tuhh.hbubert.MacFir.c
/api/v0.1/workspace/1234/src-gen:psex:c:pshdl_de_tuhh_hbubert_MacFir_sim.h
/api/v0.1/workspace/1234/src-gen:psex:c:pshdl_generic_sim.h`

				mux.HandleFunc("/api/v0.1/compiler/1234/psex/c", func(w http.ResponseWriter, r *http.Request) {
					So(r.Method, ShouldEqual, "POST")

					r.ParseForm()
					So(r.Form, ShouldResemble, url.Values{
						"module": []string{moduleName},
					})

					http.Error(w, dlURLs, http.StatusCreated)
				})

				uris, err := client.Compiler.RequestSimCode(SimC, moduleName)
				So(err, ShouldBeNil)
				So(len(uris), ShouldEqual, 3)
				// So(url, ShouldEqual, dlURL) //TODO remove the newline from the returned url
			})

			Convey("with a broken workspace, should return an error and empty url", func() {

				mux.HandleFunc("/api/v0.1/compiler/1234/psex/c", func(w http.ResponseWriter, r *http.Request) {
					So(r.Method, ShouldEqual, "POST")

					http.Error(w, `[{}]`, http.StatusBadRequest)
				})

				uris, err := client.Compiler.RequestSimCode(SimC, "abc")
				So(err, ShouldNotBeNil)
				So(uris, ShouldBeNil)
			})
		})

		Reset(teardown)
	})
}
