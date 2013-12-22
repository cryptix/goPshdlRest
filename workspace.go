package goPshdlRest

import (
	"fmt"
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
	// todo:not using NewRequest because create dosnt support json..
	// req, err := s.client.NewRequest("POST", "workspace", nil)
	rel, err := url.Parse("workspace")
	if err != nil {
		return nil, nil, err
	}

	u := s.client.BaseURL.ResolveReference(rel)

	param := url.Values{}
	param.Set("name", defaultName)
	param.Set("eMail", defaultEmail)

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(param.Encode()))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// todo /\
	//      ||

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
