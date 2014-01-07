package goPshdlRest

import (
	"fmt"
)

// CompilerService handles communication with the compiler related
// methods of the PsHdl REST API.
type CompilerService struct {
	client *Client
}

func (s *CompilerService) Validate(id string) error {
	req, err := s.client.NewRequest("POST", fmt.Sprintf("compiler/%s/validate", id), nil)
	if err != nil {
		return err
	}

	_, _, err = s.client.DoPlain(req)
	if err != nil {
		return err
	}

	return nil
}
