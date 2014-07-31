package pshdlApi

import "fmt"

// Problem is a result of a workspace validation with error describtions and solution hints
type Problem struct {
	Advise struct {
		Explanation string   `json:"explanation"`
		Message     string   `json:"message"`
		Solutions   []string `json:"solutions"`
	} `json:"advise"`
	ErrorCode string `json:"errorCode"`
	Location  struct {
		Length       float64 `json:"length"`
		Line         float64 `json:"line"`
		OffsetInLine float64 `json:"offsetInLine"`
		TotalOffset  float64 `json:"totalOffset"`
	} `json:"location"`
	Pid      float64 `json:"pid"`
	Severity string  `json:"severity"`
}

// Record desribes where a File is stored and some information about it
type Record struct {
	FileURI      string  `json:"fileURI"`
	Hash         string  `json:"hash"`
	LastModified float64 `json:"lastModified"` //TODO Float?!
	RelPath      string  `json:"relPath"`
}

// ModuleInfos describes ports and names of a module
type ModuleInfos struct {
	Instances []string `json:"instances"`
	Name      string   `json:"name"`
	Ports     []Port   `json:"ports"`
	Type      string   `json:"type"`
}

func (mi ModuleInfos) String() (s string) {
	s = fmt.Sprintf("Module[%s] - %s\n", mi.Type, mi.Name)

	s += fmt.Sprintln("Instances:", mi.Instances)

	s += fmt.Sprintln("Ports:")
	for i, port := range mi.Ports {
		s += fmt.Sprintf("#%2d [%-10s] <%2.0f bits>%-10s\n", i, port.Dir, port.Width, port.Name)
	}
	return
}

type Port struct {
	Annotations []string      `json:"annotations"`
	Dimensions  []interface{} `json:"dimensions"`
	Dir         string        `json:"dir"`
	Name        string        `json:"name"`
	Primitive   string        `json:"primitive"`
	Width       float64       `json:"width"`
}

type ByDir []Port

func (a ByDir) Len() int           { return len(a) }
func (a ByDir) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDir) Less(i, j int) bool { return a[i].Dir < a[j].Dir }

type ByName []Port

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// File describes the current state of a pshdl file in a workspace
type File struct {
	Info struct {
		Created  float64   `json:"created"`
		Files    []Record  `json:"files"`
		Problems []Problem `json:"problems"`
	} `json:"info"`
	ModuleInfos []ModuleInfos `json:"moduleInfos"`
	Record      Record        `json:"record"`
	Syntax      string        `json:"syntax"`
	Type        string        `json:"type"`
}

// Workspace represents a workspace on the API
type Workspace struct {
	Files          []File  `json:"files"`
	ID             string  `json:"id"`
	JSONVersion    string  `json:"jsonVersion"`
	LastValidation float64 `json:"lastValidation"`
	Validated      bool    `json:"validated"`
}
