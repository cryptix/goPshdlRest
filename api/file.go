package pshdlApi

type Problem struct {
	Advise struct {
		Explanation string
		Message     string
		Solutions   []string
	}
	Location struct {
		Length, Line, OffsetInLine, TotalOffset int
	}
	Pid       int
	ErrorCode string
	Severity  string
	Syntax    bool
}

type Record struct {
	RelPath      string
	FileURI      string
	LastModified int
	Hash         string
}

type ModuleInfos struct{}

type File struct {
	Syntax      string
	Type        string
	Record      Record
	ModuleInfos []ModuleInfos
	Info        struct {
		Created  int
		Problems []Problem
	}
}
