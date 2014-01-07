package goPshdlRest

import (
	// "encoding/json"
	"fmt"
	"net/http"
	// "os"

	// "github.com/cryptix/goSSEClient"
)

// StreamingService handles communication with the streaming related
// methods of the PsHdl REST API.
type StreamingService struct {
	// wrapped http client
	client *Client
	// current workspace Id
	Id string
}

type StreamingEvent interface {
	GetSubject() string
	GetFiles() []Record
	DownloadFiles() error
}

func (s *StreamingService) OpenEventStream() (*http.Response, error) {
	req, err := s.client.NewRequest("GET", "streaming/workspace/"+s.Id+"/clientID", nil)
	if err != nil {
		return nil, err
	}

	_, resp, err := s.client.DoPlain(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("TODO: Stream events")
	// url := fmt.Sprintf("http://%s/api/v0.1/streaming/workspace/%s/%s/sse", ApiHost, wp.Id, string(buf))
	// // fmt.Printf("Debug: streamingUrl:%s\n", url)

	// sseEvent, err := goSSEClient.OpenSSEUrl(url)
	// if err != nil {
	//
	// 	return err
	// }

	// wp.Events = make(chan PshdlApiStreamingEvent)

	// go func() {
	// 	for ev := range sseEvent {
	// 		var peek struct {
	// 			Subject string
	// 			MsgType string
	// 		}

	// 		err := json.Unmarshal(ev.Data, &peek)
	// 		if err != nil {
	// 			fmt.Fprintf(os.Stderr, "Error during Peek Unmarshal:%s\n", err)
	// 			done <- false
	// 		}

	// 		var apiEvent PshdlApiStreamingEvent

	// 		switch peek.Subject {

	// 		case "P:COMPILER:VHDL":
	// 			apiEvent = new(PshdlApiCompilerVhdlEvent)

	// 		case "P:COMPILER:C":
	// 			apiEvent = new(PShdlApiCompilerCEvent)

	// 		case "P:WORKSPACE:ADDED":
	// 			apiEvent = new(PshdlApiWorskpaceUpdatedEvent)

	// 		case "P:WORKSPACE:UPDATED":
	// 			apiEvent = new(PshdlApiWorskpaceUpdatedEvent)

	// 		case "P:WORKSPACE:DELETED":
	// 			apiEvent = new(PshdlApiWorskpaceDeletedEvent)

	// 		default:
	// 			fmt.Fprintf(os.Stderr, "Error unhandeld event type!:%v\n", peek)
	// 		}

	// 		err = json.Unmarshal(ev.Data, &apiEvent)
	// 		if err != nil {
	// 			fmt.Fprintf(os.Stderr, "Error during Event Unmarshal:%s\n", err)
	// 			done <- false
	// 		} else {
	// 			wp.Events <- apiEvent
	// 		}

	// 	}
	// 	fmt.Fprintln(os.Stderr, "SSEvent chan was closed.")
	// 	done <- true
	// }()

	return nil, nil
}
