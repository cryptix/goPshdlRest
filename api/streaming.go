package pshdlApi

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cryptix/goSSEClient"
)

// StreamingService handles communication with the streaming related
// methods of the PsHdl REST API.
type StreamingService struct {
	// wrapped http client
	client *Client
	// current workspace Id
	ID string
}

type StreamingEvent interface {
	GetSubject() string
	GetFiles() []Record
}

func (s *StreamingService) OpenEventStream() (<-chan StreamingEvent, error) {
	req, err := s.client.NewRequest("GET", fmt.Sprintf("streaming/workspace/%s/clientID", s.ID), nil)
	if err != nil {
		return nil, err
	}

	clientID, _, err := s.client.DoPlain(req)
	if err != nil {
		return nil, err
	}

	req, err = s.client.NewRequest("GET", fmt.Sprintf("streaming/workspace/%s/%s/sse", s.ID, string(clientID)), nil)
	if err != nil {
		return nil, err
	}

	sseEvent, err := goSSEClient.OpenSSEUrl(req.URL.String())
	if err != nil {
		return nil, err
	}

	events := make(chan StreamingEvent)

	go func() {
		for ev := range sseEvent {
			var peek struct {
				Subject string
				MsgType string
			}

			err := json.Unmarshal(ev.Data, &peek)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Peek Unmarshal:%s\n", err)
				close(events)
			}

			var apiEvent StreamingEvent

			switch peek.Subject {

			case "P:COMPILER:VHDL":
				apiEvent = new(CompilerVhdlEvent)

			case "P:COMPILER:C":
				apiEvent = new(CompilerCEvent)

			case "P:WORKSPACE:ADDED":
				apiEvent = new(WorskpaceUpdatedEvent)

			case "P:WORKSPACE:UPDATED":
				apiEvent = new(WorskpaceUpdatedEvent)

			case "P:WORKSPACE:DELETED":
				apiEvent = new(WorskpaceDeletedEvent)

			default:
				fmt.Fprintf(os.Stderr, "Error unhandeld event type!:%v\n", peek)
			}

			err = json.Unmarshal(ev.Data, &apiEvent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\ndata:%s\n", err, ev.Data)
				close(events)
			} else {
				events <- apiEvent
			}

		}
		fmt.Fprintln(os.Stderr, "SSEvent chan was closed.")
		close(events)
	}()

	return events, nil

}
