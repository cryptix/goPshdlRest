package goPshdlRest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultName  = "JohnGo"
	defaultEmail = "none@me.com"
)

// WorkspaceService handles communication with the workspace related
// methods of the PsHdl REST API.
type WorkspaceService struct {
	// wrapped http client
	client *Client
	// current workspace ID
	ID string
}

// Workspace represents a workspace on the API
type Workspace struct {
	ID             string
	Files          []File
	LastValIDation int
	ValIDated      bool
	JSONVersion    string `json:"JsonVersion"`
}

// Create creates a new Workspace on the Rest API
// Currently using form encoded post, want json..!
func (s *WorkspaceService) Create() (*Workspace, *http.Response, error) {
	// prepare request
	param := url.Values{}
	param.Set("name", defaultName)
	param.Set("eMail", defaultEmail)

	req, err := s.client.NewReaderRequest("POST", "workspace", strings.NewReader(param.Encode()), "")
	if err != nil {
		return nil, nil, err
	}

	// do the request
	body, resp, err := s.client.DoPlain(req)
	if err != nil {
		return nil, resp, err
	}

	wsCreatedRegex := regexp.MustCompile(`/api/v0.1/workspace/([0-9A-F]*)`)
	matches := wsCreatedRegex.FindSubmatch(body)
	if len(matches) != 2 {
		return nil, resp, fmt.Errorf("no Workspace ID - %s", string(body))
	}

	s.ID = string(matches[1])
	w := &Workspace{
		ID: s.ID,
	}

	return w, resp, nil
}

// GetInfo gets all the info there is to get for a PSHDL Workspace
func (s *WorkspaceService) GetInfo() (*Workspace, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "workspace/"+s.ID, nil)
	if err != nil {
		return nil, nil, err
	}

	w := new(Workspace)
	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, resp, err
	}

	if w.ID != s.ID {
		return nil, nil, fmt.Errorf("we got response for %v a different workspace", w)
	}

	return w, resp, err
}

// Delete removes the file `fname` from the specified workspace
func (s *WorkspaceService) Delete(fname string) (bool, *http.Response, error) {
	if s.ID == "" {
		return false, nil, fmt.Errorf("workspace ID not set")
	}

	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("workspace/%s/%s", s.ID, fname), nil)
	if err != nil {
		return false, nil, err
	}

	_, resp, err := s.client.DoPlain(req)
	if err != nil {
		return false, resp, err
	}

	if resp.StatusCode != 200 {
		return false, resp, fmt.Errorf("file was not deleted")
	}

	return true, resp, err
}

// UploadFile adds a file with fname to the Workspace specified by ID
func (s *WorkspaceService) UploadFile(fname string, fbuf io.Reader) error {
	if s.ID == "" {
		return fmt.Errorf("workspace ID not set")
	}

	// convert Upload to Multipart
	reqBody := &bytes.Buffer{}
	writer := multipart.NewWriter(reqBody)

	part, err := writer.CreateFormFile("file", fname)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, fbuf)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	// prepare request
	req, err := s.client.NewReaderRequest("POST", fmt.Sprintf("workspace/%s", s.ID), reqBody, writer.FormDataContentType())
	if err != nil {
		return err
	}

	// do the request
	_, _, err = s.client.DoPlain(req)
	if err != nil {
		return err
	}

	return nil
}

func (s *WorkspaceService) DownloadAllFiles() error {
	wp, _, err := s.GetInfo()
	if err != nil {
		return fmt.Errorf("s.GetInfo() Error: %s", err)
	}

	errc := make(chan error)
	fileCount := len(wp.Files)

	if fileCount == 0 {
		return nil
	}

	start := time.Now()

	for _, f := range wp.Files {
		go func(f File) {
			err := s.DownloadFile(f.Record.RelPath)
			if err != nil {
				errc <- err //fmt.Fprintf(os.Stderr, "Could not http.Get %s - %s\n", file.Record.RelPath, err)
				return
			}
			errc <- nil
		}(f)
	}

	for {
		select {
		case err := <-errc:
			if err != nil {
				close(errc)
				return fmt.Errorf("could not load all files. Error: %s", err)
			}

			fileCount--
			if fileCount == 0 {
				return nil
			}

		case <-time.After(5 * time.Second):
			fmt.Fprintf(os.Stderr, "DownloadAllFiles() Waiting.. \n%d files left. Duration: %s\n", fileCount, time.Since(start))
		}
	}
}

// DownloadFile returns a copy of fname
func (s *WorkspaceService) DownloadFile(fname string) error {
	if s.ID == "" {
		return fmt.Errorf("workspace ID not set")
	}

	req, err := s.client.NewRequest("GET", fmt.Sprintf("workspace/%s/%s", s.ID, fname), nil)
	if err != nil {
		return fmt.Errorf("client.NewRequest() error: %s", err)
	}

	req.Header.Set("Accept", "text/plain")

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return fmt.Errorf("client.Do(req) error: %s", err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return fmt.Errorf("os.OpenFile() error: %s", err)
	}
	defer f.Close()

	copied, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	fileLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	if int(copied) != fileLength {
		return fmt.Errorf(" io.Copy(f, resp.Body) did not copy the whole file. got <%d> wanted <%d>", copied, fileLength)
	}

	return nil
}
