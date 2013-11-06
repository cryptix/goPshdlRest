package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/cryptix/goSSEClient"
)

type PshdlApiStreamingEvent interface {
	GetSubject() string
	GetFiles() []PshdlApiRecord
}

func (wp *PshdlWorkspace) OpenEventStream(done chan bool) error {
	// todo we need a unique client id!
	url := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/%d/sse", ApiHost, wp.Id, rand.Intn(128))

	sseEvent, err := goSSEClient.OpenSSEUrl(url)
	if err != nil {
		done <- false
		return err
	}

	wp.Events = make(chan PshdlApiStreamingEvent)

	go func() {
		for ev := range sseEvent {
			var peek struct {
				Subject string
				MsgType string
			}

			err := json.Unmarshal(ev.Data, &peek)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Peek Unmarshal:%s\n", err)
				done <- false
			}

			var apiEvent PshdlApiStreamingEvent

			switch peek.Subject {

			case "P:COMPILER:VHDL":
				apiEvent = new(PshdlApiCompilerVhdlEvent)

			case "P:COMPILER:C":
				apiEvent = new(PShdlApiCompilerCEvent)

			case "P:WORKSPACE:ADDED":
				apiEvent = new(PshdlApiWorskpaceUpdatedEvent)

			case "P:WORKSPACE:UPDATED":
				apiEvent = new(PshdlApiWorskpaceUpdatedEvent)

			case "P:WORKSPACE:DELETED":
				apiEvent = new(PshdlApiWorskpaceDeletedEvent)

			default:
				fmt.Fprintf(os.Stderr, "Error unhandeld event type!:%v\n", peek)
				done <- false
			}

			err = json.Unmarshal(ev.Data, &apiEvent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\n", err)
				done <- false
			} else {
				wp.Events <- apiEvent
			}

		}
		fmt.Fprintln(os.Stderr, "SSEvent chan was closed.")
		done <- true
	}()

	return nil
}
