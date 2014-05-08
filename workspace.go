package goPshdlRest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

// DownloadFile returns a copy of fname
func (s *WorkspaceService) DownloadFile(fname string) ([]byte, error) {
	if s.ID == "" {
		return nil, fmt.Errorf("workspace ID not set")
	}

	req, err := s.client.NewRequest("GET", fmt.Sprintf("workspace/%s/%s", s.ID, fname), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := s.client.DoPlain(req)
	if err != nil {
		return nil, err
	}

	return body, nil
}
