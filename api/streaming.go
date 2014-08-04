package pshdlApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cryptix/goSSEClient"
)

// StreamingService handles communication with the streaming related
// methods of the PsHdl REST API.
type StreamingService struct {
	// wrapped http client
	client *Client
	// current workspace Id
	ID       string
	clientID string
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

	cid, _, err := s.client.DoPlain(req)
	if err != nil {
		return nil, err
	}
	s.clientID = string(cid)
	dbg("OpenEventStream() Client ID:%s", s.clientID)

	req, err = s.client.NewRequest("GET", fmt.Sprintf("streaming/workspace/%s/%s/sse", s.ID, s.clientID), nil)
	if err != nil {
		return nil, err
	}

	sseEvent, err := goSSEClient.OpenSSEUrl(req.URL.String())
	if err != nil {
		return nil, err
	}
	dbg("OpenEventStream sseEvent channel open")

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

			case "P:PING":
				apiEvent = new(PingEvent)

			default:
				fmt.Fprintf(os.Stderr, "Error unhandeld event type!:%v\n", peek)
				continue
			}

			dbg("ssEvent: %s", peek.Subject)

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

type StreamingClientEvent struct {
	ID        string `json:"clientID"`
	Timestamp int64  `json:"timeStamp"`
	Subject   string `json:"subject"`
}

func (s *StreamingService) SendClientConnected() error {
	dbg("Streaming.SendClientConnected(%s) Client:%s", s.ID, s.clientID)

	body, err := json.Marshal(StreamingClientEvent{
		ID:        s.clientID,
		Timestamp: time.Now().Unix(),
		Subject:   "P:CLIENT:CONNECTED",
	})
	if err != nil {
		return err
	}

	req, err := s.client.NewReaderRequest(
		"POST",
		fmt.Sprintf("streaming/workspace/%s/%s", s.ID, s.clientID),
		bytes.NewReader(body),
		"application/json",
	)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return nil
	}

	if err = CheckResponse(resp); err != nil {
		return err
	}

	return nil
}
