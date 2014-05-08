package goPshdlRest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(want, r.Form) {
		t.Errorf("Request parameters = %v, want %v", r.Form, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if value := r.Header.Get(header); want != value {
		t.Errorf("Header %s = %s, want: %s", header, value, want)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Unable to read body")
	}
	str := string(b)
	if want != str {
		t.Errorf("Body = %s, want: %s", str, want)
	}
}

// Helper function to test that a value is marshalled to JSON as expected.
func testJSONMarshal(t *testing.T, v interface{}, want string) {
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %v", v)
	}

	w := new(bytes.Buffer)
	err = json.Compact(w, []byte(want))
	if err != nil {
		t.Errorf("String is not valid json: %s", want)
	}

	if w.String() != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}

	// now go the other direction and make sure things unmarshal as expected
	u := reflect.ValueOf(v).Interface()
	if err := json.Unmarshal([]byte(want), u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v", want)
	}

	if !reflect.DeepEqual(v, u) {
		t.Errorf("json.Unmarshal(%q) returned %s, want %s", want, u, v)
	}
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
