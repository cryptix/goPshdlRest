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
	client *Client
}

// Workspace represents a workspace on the API
type Workspace struct {
	Id             string
	Files          []File
	LastValidation int
	Validated      bool
	JsonVersion    string
}

// Creates a new Workspace on the Rest API
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
		return nil, resp, fmt.Errorf("No Workspace ID - %s", string(body))
	}

	w := &Workspace{
		Id: string(matches[1]),
	}

	return w, resp, nil
}

// GetInfo gets all the info there is to get for a PSHDL Workspace
func (s *WorkspaceService) GetInfo(id string) (*Workspace, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "workspace/"+id, nil)
	if err != nil {
		return nil, nil, err
	}

	w := new(Workspace)
	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, resp, err
	}

	if w.Id != id {
		return nil, nil, fmt.Errorf("We got response for a different workspace!!", w)
	}

	return w, resp, err
}

// Delete's the file `fname` from the specified workspace
func (s *WorkspaceService) Delete(id, fname string) (bool, *http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("workspace/%s/%s", id, fname), nil)
	if err != nil {
		return false, nil, err
	}

	_, resp, err := s.client.DoPlain(req)
	if err != nil {
		return false, resp, err
	}

	if resp.StatusCode != 200 {
		return false, resp, fmt.Errorf("File was not deleted.")
	}

	return true, resp, err
}

// Uploads a file with fname to the Workspace specified by id
func (s *WorkspaceService) UploadFile(id, fname string, fbuf io.Reader) error {

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
	req, err := s.client.NewReaderRequest("POST", fmt.Sprintf("workspace/%s/%s", id, fname), reqBody, writer.FormDataContentType())
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
