package goPshdlRest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = NewClient(nil)
	url, _ := url.Parse(server.URL + "/api/v0.1/")
	client.BaseURL = url
	client.Workspace.Id = "1234"
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	var c *Client
	Convey("Given a new Client", t, func() {
		c = NewClient(nil)

		Convey("It should have the correct BaseURL", func() {
			So(c.BaseURL.String(), ShouldEqual, defaultBaseURL)
		})

		Convey("It should have the correct UserAgent", func() {
			So(c.UserAgent, ShouldEqual, userAgent)
		})
	})
}

func TestNewRequest(t *testing.T) {
	var (
		c   *Client
		req *http.Request
	)

	type createPut struct {
		Name, Email string
	}

	Convey("Given a new Client", t, func() {
		c = NewClient(nil)

		Convey("and a valid Request", func() {
			inURL, outURL := "foo", defaultBaseURL+"foo"
			inBody, outBody := &createPut{Name: "l", Email: "hi@me.com"}, `{"Name":"l","Email":"hi@me.com"}`+"\n"
			req, _ = c.NewRequest("PUT", inURL, inBody)

			Convey("It should have its URL expanded", func() {
				So(req.URL.String(), ShouldEqual, outURL)
			})

			Convey("It should encode the body in JSON", func() {
				body, _ := ioutil.ReadAll(req.Body)
				So(string(body), ShouldEqual, outBody)
			})

			Convey("It should have the default user-agent is attached to the request", func() {
				userAgent := req.Header.Get("User-Agent")
				So(c.UserAgent, ShouldEqual, userAgent)
			})

		})

		Convey("and an invalid Request", func() {
			type T struct {
				A map[int]interface{}
			}
			_, err := c.NewRequest("GET", "/", &T{})

			Convey("It should return an error (beeing *json.UnsupportedTypeError)", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, &json.UnsupportedTypeError{})
			})

		})

		Convey("and a bad Request URL", func() {
			_, err := c.NewRequest("GET", ":", nil)
			Convey("It should return an error (beeing *url.Error{})", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, &url.Error{})
			})
		})
	})

}

func TestDo(t *testing.T) {

	Convey("Given a clean test server", t, func() {
		setup()

		Convey("Do() should send the request", func() {

			type foo struct {
				A string
			}

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				So(r.Method, ShouldEqual, "GET")

				fmt.Fprint(w, `{"A":"a"}`)
			})

			req, _ := client.NewRequest("GET", "/", nil)
			body := new(foo)
			client.Do(req, body)

			So(body, ShouldResemble, &foo{"a"})
		})

		Convey("A Bad Request should return an error", func() {

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Bad Request", 400)
			})

			req, _ := client.NewRequest("GET", "/", nil)
			_, err := client.Do(req, nil)
			So(err, ShouldNotBeNil)

		})

		Convey("A plain request should get response", func() {

			want := `/api/v0.1/servertime`

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				So(r.Method, ShouldEqual, "GET")
				fmt.Fprint(w, want)
			})

			req, _ := client.NewRequest("GET", "/", nil)
			resp, _, _ := client.DoPlain(req)

			body := string(resp)
			So(body, ShouldEqual, want)
		})

		Convey("A bad plain request should return a http error", func() {

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Bad Request", 400)
			})

			req, _ := client.NewRequest("GET", "/", nil)
			_, _, err := client.DoPlain(req)
			So(err, ShouldNotBeNil)
		})

		Reset(teardown)
	})

}
