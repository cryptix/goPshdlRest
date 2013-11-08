package main

type PshdlEventMetaInfo struct {
	Subject   string
	MsgType   string
	TimeStamp int
}

// P:WORKSPACE:ADDED
// P:WORKSPACE:UPDATED
type PshdlApiWorskpaceUpdatedEvent struct {
	PshdlEventMetaInfo
	Contents []PshdlApiFile
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiWorskpaceUpdatedEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, len(ev.Contents))
	for i, f := range ev.Contents {
		files[i] = f.Record
	}
	return files
}

// P:WORKSPACE:DELETED
type PshdlApiWorskpaceDeletedEvent struct {
	PshdlEventMetaInfo
	Contents PshdlApiFile
}

func (ev *PshdlApiWorskpaceDeletedEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiWorskpaceDeletedEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, 1)
	files[0] = ev.Contents.Record
	return files
}

// P:COMPILER:VHDL
type PshdlApiCompilerVhdlEvent struct {
	PshdlEventMetaInfo
	Contents []struct {
		Created  int
		Problems []PshdlApiProblem
		Files    []PshdlApiRecord
	}
}

func (ev *PshdlApiCompilerVhdlEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PshdlApiCompilerVhdlEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, len(ev.Contents))
	for i, f := range ev.Contents {
		if len(f.Files) == 1 {
			files[i] = f.Files[0]
		}
	}
	return files
}

// P:COMPILER:C
type PShdlApiCompilerCEvent struct {
	PshdlEventMetaInfo
	Contents PshdlApiRecord
}

func (ev *PShdlApiCompilerCEvent) GetSubject() string {
	return ev.Subject
}

func (ev *PShdlApiCompilerCEvent) GetFiles() []PshdlApiRecord {
	files := make([]PshdlApiRecord, 1)
	files[0] = ev.Contents
	return files
}
