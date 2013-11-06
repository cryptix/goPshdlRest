package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ApiHost      = "api.pshdl.org"
	workspaceUrl = "/api/v0.1/workspace/"
)

// POST
// with name and email
type PshdlWorkspace struct {
	Id, Name, Email string
	Files           []PshdlApiFile
	LastValidation  int
	Validated       bool
	Events          chan PshdlApiStreamingEvent
}

func (wp PshdlWorkspace) String() string {
	if wp.Id == "" {
		return fmt.Sprintf("http://%s%s", ApiHost, workspaceUrl)
	}
	return fmt.Sprintf("http://%s%s%s", ApiHost, workspaceUrl, wp.Id)
}

type PshdlApiFile struct {
	Syntax string
	Type   string
	Record PshdlApiRecord
	Info   struct {
		Created  int
		Problems []PshdlApiProblem
	}
}

type PshdlApiRecord struct {
	RelPath      string
	FileURI      string
	LastModified int
}

func (file PshdlApiFile) String() (str string) {
	str = fmt.Sprintf("\nName: %s - ", file.Record.RelPath)
	if len(file.Info.Problems) > 0 {
		str += fmt.Sprintf("Problems:\n")
		for _, prob := range file.Info.Problems {
			str += fmt.Sprintf("%4d: %s\n", prob.Location.Line, prob.ErrorCode)
		}
	} else {
		str += fmt.Sprintf("No Problems.")
	}

	return
}

type PshdlApiProblem struct {
	Advise struct {
		Explanation string
		Message     string
		Solutions   []string
	}
	Location struct {
		Length, Line, OffsetInLine, TotalOffset int
	}
	Pid       int
	ErrorCode string
	Severity  string
	Syntax    bool
}

func OpenWorkspace(id string) (*PshdlWorkspace, error) {
	wp := &PshdlWorkspace{Id: id}
	if err := wp.parseWorkspace(); err != nil {
		return nil, err
	}

	return wp, nil
}

func NewWorkspace(name, email string) (*PshdlWorkspace, error) {
	wp := &PshdlWorkspace{Name: name, Email: email}
	if err := wp.createWorkspace(); err != nil {
		return nil, err
	}
	return wp, nil
}

func (wp *PshdlWorkspace) AddFile(path string) error {
	if wp.Id == "" {
		return fmt.Errorf("Workspace not Open.\n")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fname := filepath.Base(path)
	part, err := writer.CreateFormFile("file", fname)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", wp.String(), body)
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("Could not save %s - %s\n", fname, resp.StatusCode)
	}

	return nil
}

func (wp *PshdlWorkspace) Validate() error {
	if wp.Id == "" {
		return fmt.Errorf("Workspace not Open.\n")
	}

	url := fmt.Sprintf("http://%s/api/v0.1/compiler/%s/validate", ApiHost, wp.Id)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	wp.parseApiRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (wp *PshdlWorkspace) DownloadAllFiles() error {

	// probe folder
	err := os.Mkdir(wp.Id, 0700)
	if err != nil {
		return err
	}

	done := make(chan bool)
	fileCount := len(wp.Files)

	if fileCount == 0 {
		return nil
	}

	start := time.Now()

	for _, file := range wp.Files {
		go func(file PshdlApiFile) {
			url := fmt.Sprintf("http://%s%s", ApiHost, file.Record.FileURI)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Accept", "text/plain")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not http.Get %s - %s\n", file.Record.RelPath, err)
				done <- false
				return
			}
			defer resp.Body.Close()

			fname := fmt.Sprintf("%s/%s", wp.Id, file.Record.RelPath)
			out, err := os.Create(fname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not os.Create file %s - %s\n", file.Record.RelPath, err)
				done <- false
				return
			}

			io.Copy(out, resp.Body)
			done <- true
		}(file)
	}

	for {
		select {
		case worked := <-done:
			if worked == false {
				close(done)
				return fmt.Errorf("Could not load all files.")
			}

			fileCount -= 1
			if fileCount == 0 {
				return nil
			}

		case <-time.After(3 * time.Second):
			fmt.Fprintf(os.Stderr, "Waiting.. %d files left. Duration: %s\n", fileCount, time.Since(start))
		}
	}
}

func (wp *PshdlWorkspace) createWorkspace() error {
	param := url.Values{}
	param.Set("name", wp.Name)
	param.Set("eMail", wp.Email)

	resp, err := http.PostForm(wp.String(), param)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("Workspace was not created - StatusCode:%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	out := string(body)

	if !strings.HasPrefix(out, workspaceUrl) {
		return fmt.Errorf("No Workspace ID - %s", out)
	}
	wp.Id = out[len(workspaceUrl):]
	return nil
}

func (wp *PshdlWorkspace) parseWorkspace() error {
	req, err := http.NewRequest("GET", wp.String(), nil)
	req.Header.Set("Accept", "application/json")

	err = wp.parseApiRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (wp *PshdlWorkspace) parseApiRequest(req *http.Request) error {
	// fmt.Printf("parse Request:%+v\n", req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("Workspace or Resource not found")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response Status %d", resp.StatusCode)
	}

	// fmt.Printf("parse Response:%+v\n", resp)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&wp)
	if err != nil {
		return err
	}
	return nil
}
