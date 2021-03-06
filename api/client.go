package pshdlApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/visionmedia/go-debug"
)

var dbg = debug.Debug("pshdlApi")

const (
	libraryVersion = "0.1"
	defaultBaseURL = "http://api6.pshdl.org/api/v0.1/"
	userAgent      = "pshdlApi/" + libraryVersion

	defaultAccept    = "application/json"
	defaultMediaType = "application/octet-stream"
)

// A Client manages communication with the Pshdl Rest API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests.  baseURL should always be specified with a trailing slash.
	baseURL *url.URL

	// User agent used when communicating with the PSHDL REST API.
	UserAgent string

	// Services used for talking to different parts of the PSHDL REST API.
	Workspace *WorkspaceService
	Compiler  *CompilerService
	Streaming *StreamingService
}

// NewClient returns a new PSHDL REST API client.  If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	defer dbg("NewClient()")

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		panic(err)
	}

	c := &Client{client: httpClient, baseURL: baseURL, UserAgent: userAgent}
	c.Workspace = &WorkspaceService{client: c}
	c.Compiler = &CompilerService{client: c}
	c.Streaming = &StreamingService{client: c}

	return c
}

// NewClientWithID returns a new PSHDL REST API client.  If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClientWithID(httpClient *http.Client, id string) *Client {
	defer dbg("NewClientWithID(%s)", id)

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		panic(err)
	}

	c := &Client{client: httpClient, baseURL: baseURL, UserAgent: userAgent}
	c.Workspace = &WorkspaceService{client: c, ID: id}
	c.Compiler = &CompilerService{client: c, ID: id}
	c.Streaming = &StreamingService{client: c, ID: id}

	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the baseURL of the Client.
// Relative URLs should always be specified without a preceding slash.  If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (req *http.Request, err error) {
	defer dbg("client.NewRequest[%s]: %s - %v", method, urlStr, err)

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err = http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", defaultAccept)
	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// NewReaderRequest creates an API request. Uses a io.Reader and ctype instead of marshaling json.
func (c *Client) NewReaderRequest(method, urlStr string, body io.Reader, ctype string) (req *http.Request, err error) {
	defer dbg("client.NewReaderRequest[%s] %s - %v", method, urlStr, err)

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	req, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "text/plain")
	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", ctype)
	if ctype == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

// Do sends an API request and returns the API response.  The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (resp *http.Response, err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			dbg("client.Do(%s) Error! (%s)", req.URL.Path, err)
			return
		}
		dbg("client.Do(%s) %s (%v)", req.URL.Path, resp.Status, time.Since(start))
	}()

	resp, err = c.client.Do(req)
	if err != nil {
		return nil, err
	}

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		resp.Body.Close()
	}
	return resp, err
}

// DoPlain sends an API request and returns the API response as a slice of bytes.
func (c *Client) DoPlain(req *http.Request) (data []byte, resp *http.Response, err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			dbg("client.DoPlain(%s) Error! (%s)", req.URL.Path, err)
			return
		}
		dbg("client.DoPlain(%s) %s (%v)", req.URL.Path, resp.Status, time.Since(start))
	}()

	req.Header.Set("Accept", "text/plain")

	resp, err = c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return nil, resp, err
	}

	data, err = ioutil.ReadAll(resp.Body)
	return data, resp, err
}

/*
An ErrorResponse reports one or more errors caused by an API request.

PSHDL REST API docs: http://developer.github.com/v3/#client-errors
*/
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  interface{}
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.  API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.  Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}
