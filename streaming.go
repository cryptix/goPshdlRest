package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cryptix/goSSEClient"
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

func (wp *PshdlWorkspace) OpenEventStream(done chan bool) error {
	// todo we need a unique client id!
	url := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/1/sse", ApiHost, wp.Id)

	sseEvent, err := goSSEClient.OpenSSEUrl(url)
	if err != nil {
		done <- false
		return err
	}

	wp.Events = make(chan PshdlApiStreamingEvent)

	go func() {
		for ev := range sseEvent {
			// doenst work since we need to unmarshal data twice depending on MsgType..
			// dec := json.NewDecoder(&ev.Data)
			// err := dec.Unmarshal(&ApiEvent)

			var peek struct {
				Subject string
				MsgType string
			}
			data := ev.Data.Bytes()

			err := json.Unmarshal(data, &peek)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Peek Unmarshal:%s\n", err)
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

			err = json.Unmarshal(data, &ev)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\n", err)
				done <- false
			} else {
				wp.Events <- ev
			}

		}
		fmt.Fprintln(os.Stderr, "SSEvent chan was closed.")
		done <- true
	}()

	return nil
}
