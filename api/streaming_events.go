package pshdlApi

import (
	"fmt"
	"os"
)

type PshdlEventMetaInfo struct {
	Subject   string
	MsgType   string
	TimeStamp int
}

// P:WORKSPACE:ADDED
// P:WORKSPACE:UPDATED
type WorskpaceUpdatedEvent struct {
	PshdlEventMetaInfo
	Contents []File
}

func (ev *WorskpaceUpdatedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *WorskpaceUpdatedEvent) GetFiles() []Record {
	files := make([]Record, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Record
	}
	return files
}

func (ev *WorskpaceUpdatedEvent) DownloadFiles(ws WorkspaceService) error {
	fmt.Fprintln(os.Stderr, "[!] Download not support for WorskpaceUpdatedEvent.")
	return nil
}

// P:WORKSPACE:DELETED
type WorskpaceDeletedEvent struct {
	PshdlEventMetaInfo
	Contents File
}

func (ev *WorskpaceDeletedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *WorskpaceDeletedEvent) GetFiles() []Record {
	files := make([]Record, 1)
	files[0] = ev.Contents.Record
	return files
}

// P:COMPILER:VHDL
type CompilerVhdlEvent struct {
	PshdlEventMetaInfo
	Contents []struct {
		Created  int
		Problems []Problem
		Files    []Record
	}
}

func (ev *CompilerVhdlEvent) GetSubject() string {
	return ev.Subject
}

func (ev *CompilerVhdlEvent) GetFiles() []Record {
	files := make([]Record, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Files[0]
		if len(f.Files) != 1 {
			fmt.Fprintln(os.Stderr, "[!] CompilerVhdlEvent.GetFiles() Warning: Multiple Files inside ContentsRecord.")
		}
	}
	return files
}

// P:COMPILER:C
type CompilerCEvent struct {
	PshdlEventMetaInfo
	Contents []struct {
		Created  int
		Problems []Problem
		Files    []Record
	}
}

func (ev *CompilerCEvent) GetSubject() string {
	return ev.Subject
}

func (ev *CompilerCEvent) GetFiles() []Record {
	files := make([]Record, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Files[0]
		if len(f.Files) != 1 {
			fmt.Fprintln(os.Stderr, "[!] CompilerCEvent.GetFiles() Warning: Multiple Files inside ContentsRecord.")
		}
	}
	return files
}
