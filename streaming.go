package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type PshdlApiStreamingEvent interface {
	GetSubject() string
	GetFiles() []PshdlApiRecord
}

type PshdlApiWorskpaceUpdatedEvent struct {
	Subject   string
	MsgType   string
	TimeStamp int
	Contents  []PshdlApiFile
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Record
	}
	return files
}

type PshdlApiCompiledVhdEvent struct {
	Subject   string
	MsgType   string
	TimeStamp int
	Contents  []struct {
		Created  int
		Problems []PshdlApiProblem
		Files    []PshdlApiRecord
	}
}

func (ev *PshdlApiCompiledVhdEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiCompiledVhdEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, len(ev.Contents))
	for i, f := range ev.Contents {
		if len(f.Files) == 1 {
			files[i] = f.Files[0]
		}
	}
	return files
}

func (wp *PshdlWorkspace) OpenEventStream(events chan PshdlApiStreamingEvent, done chan bool) error {

	url := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/1/sse", ApiHost, wp.Id)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: resp.StatusCode == %d\n", resp.StatusCode)
	}

	go processStream(resp.Body, events, done)

	return nil
}

func processStream(body io.ReadCloser, events chan PshdlApiStreamingEvent, done chan bool) {
	var evBuf string

	buffed := bufio.NewReader(body)
	for {
		line, err := buffed.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during resp.Body read:%s\n", err)
			done <- false
		}

		switch {

		// start of event
		case strings.HasPrefix(line, "id:"):
			// todo: regexp
			// parts := strings.Split(line, ":")

		// event data
		case strings.HasPrefix(line, "data:"):
			evBuf += line[6:]

		// end of event
		case len(line) == 1:
			var peek struct {
				Subject string
				MsgType string
			}
			err := json.Unmarshal([]byte(evBuf), &peek)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\n", err)
				done <- false
			}

			var ev PshdlApiStreamingEvent

			switch peek.Subject {
			case "P:COMPILER:VHDL":
				ev = new(PshdlApiCompiledVhdEvent)
			case "P:WORKSPACE:UPDATED":
				ev = new(PshdlApiWorskpaceUpdatedEvent)
			default:
				fmt.Fprintf(os.Stderr, "Error unhandeld event type!:%v\n", peek)
				done <- false
			}

			err = json.Unmarshal([]byte(evBuf), ev)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\n", err)
				done <- false
			} else {
				events <- ev
			}

			// create fresh event
			evBuf = ""

		default:
			fmt.Fprintf(os.Stderr, "Error during EventReadLoop - Default triggerd! len:%d\n%s", len(line), line)
		}
	}
}
