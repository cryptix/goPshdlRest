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

func (ev *CompilerVhdlEvent) GetFiles() (rec []Record) {
	for _, f := range ev.Contents {
		for _, r := range f.Files {
			rec = append(rec, r)
		}
	}
	return
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
	if len(ev.Contents) == 0 {
		return nil
	}

	recIdx, recLen := 0, 0
	for _, f := range ev.Contents {
		recLen += len(f.Files)
	}
	records := make([]Record, recLen)

	for _, f := range ev.Contents {
		for _, rec := range f.Files {
			records[recIdx] = rec
			recIdx++
		}
	}
	return records
}

type PingEvent struct {
	PshdlEventMetaInfo
}

func (ev *PingEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PingEvent) GetFiles() []Record {
	return nil
}
