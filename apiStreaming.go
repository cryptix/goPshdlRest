package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cryptix/goSSEClient"
)

type PshdlApiStreamingEvent interface {
	GetSubject() string
	GetFiles() []PshdlApiRecord
}

func (wp *PshdlWorkspace) OpenEventStream(done chan bool) error {
	// todo we need a unique client id!
	clientIdUrl := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/clientID", ApiHost, wp.Id)
	// fmt.Printf("Debug: clientIdUrl:%s\n", clientIdUrl)

	resp, err := http.Get(clientIdUrl)
	if err != nil {
		done <- false
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		done <- false
		return fmt.Errorf("Error: ClientId response code: %d", resp.StatusCode)
	}

	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil {
		done <- false
		return fmt.Errorf("Error: ClientId response could not be read: %s", err)
	}

	url := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/%s/sse", ApiHost, wp.Id, string(buf))
	// fmt.Printf("Debug: streamingUrl:%s\n", url)

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
