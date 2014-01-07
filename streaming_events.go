package goPshdlRest

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
type PshdlApiWorskpaceUpdatedEvent struct {
	PshdlEventMetaInfo
	Contents []File
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetFiles() []Record {
	files := make([]Record, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Record
	}
	return files
}

func (ev *PshdlApiWorskpaceUpdatedEvent) DownloadFiles() error {
	fmt.Fprintln(os.Stderr, "[!] Download not support for WorskpaceUpdatedEvent.")
	return nil
}

// P:WORKSPACE:DELETED
type PshdlApiWorskpaceDeletedEvent struct {
	PshdlEventMetaInfo
	Contents File
}

func (ev *PshdlApiWorskpaceDeletedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiWorskpaceDeletedEvent) GetFiles() []Record {
	files := make([]Record, 1)
	files[0] = ev.Contents.Record
	return files
}

func (ev *PshdlApiWorskpaceDeletedEvent) DownloadFiles() error {
	fmt.Fprintln(os.Stderr, "[!] Download not support for WorskpaceDeletedEvent.")
	return nil
}

// P:COMPILER:VHDL
type PshdlApiCompilerVhdlEvent struct {
	PshdlEventMetaInfo
	Contents []struct {
		Created  int
		Problems []Problem
		Files    []Record
	}
}

func (ev *PshdlApiCompilerVhdlEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiCompilerVhdlEvent) GetFiles() []Record {
	files := make([]Record, len(ev.Contents))
	for i, f := range ev.Contents {
		if len(f.Files) == 1 {
			files[i] = f.Files[0]
		}
	}
	return files
}

func (ev *PshdlApiCompilerVhdlEvent) DownloadFiles() error {
	return downloadApiFiles(ev)
}

// P:COMPILER:C
type PShdlApiCompilerCEvent struct {
	PshdlEventMetaInfo
	Contents Record
}

func (ev *PShdlApiCompilerCEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PShdlApiCompilerCEvent) GetFiles() []Record {
	files := make([]Record, 1)
	files[0] = ev.Contents
	return files
}

func (ev *PShdlApiCompilerCEvent) DownloadFiles() error {
	return downloadApiFiles(ev)
}

func downloadApiFiles(ev PshdlApiStreamingEvent) error {
	files := ev.GetFiles()

	count := len(files)

	if count == 0 {
		fmt.Println("[*] No Files to download.")
		return nil
	}

	errc := make(chan error)

	for _, file := range files {
		fmt.Printf("[*] Downloading %s\n", file.RelPath)
		// ugly...
		go func(f Record) {
			f.DownloadFile(errc)
		}(file)
	}

	for err := range errc {
		if err != nil {
			return err
		}
		count -= 1
		if count == 0 {
			close(errc)
		}
	}
	return nil
}