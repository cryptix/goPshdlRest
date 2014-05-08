package goPshdlRest

import (
	"fmt"
	"net/http"
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

		Reset(teardown)
	})
}
