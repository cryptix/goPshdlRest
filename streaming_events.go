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

func (ev *WorskpaceUpdatedEvent) DownloadFiles() error {
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

func (ev *WorskpaceDeletedEvent) DownloadFiles() error {
	fmt.Fprintln(os.Stderr, "[!] Download not support for WorskpaceDeletedEvent.")
	return nil
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
		if len(f.Files) == 1 {
			files[i] = f.Files[0]
		}
	}
	return files
}

func (ev *CompilerVhdlEvent) DownloadFiles() error {
	return downloadApiFiles(ev)
}

// P:COMPILER:C
type CompilerCEvent struct {
	PshdlEventMetaInfo
	Contents Record
}

func (ev *CompilerCEvent) GetSubject() string {
	return ev.Subject
}

func (ev *CompilerCEvent) GetFiles() []Record {
	files := make([]Record, 1)
	files[0] = ev.Contents
	return files
}

func (ev *CompilerCEvent) DownloadFiles() error {
	return downloadApiFiles(ev)
}

func downloadApiFiles(ev StreamingEvent) error {
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
			// f.DownloadFile(errc)
			fmt.Println("Download files TODO")
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
