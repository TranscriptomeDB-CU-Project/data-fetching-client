package dtos

type Attributes struct {
	Name          string       `json:"name"`
	Value         string       `json:"value"`
	Valqualifiers []Attributes `json:"valqualifiers"`
}

type AccessionFileMetadata struct {
	Path string `json:"path"`
	Size int    `json:"size"`
}

type ResultMetadata struct {
	Name         string
	TimeModified int
}

type SubsectionMetadata struct {
	Accno       string        `json:"accno"`
	Type        string        `json:"type"`
	Attributes  []Attributes  `json:"attributes"`
	Subsections []interface{} `json:"subsections"`
	Files       []interface{} `json:"files"`
}

type SectionMetadata struct {
	Accno      string        `json:"accno"`
	Type       string        `json:"type"`
	Attributes []Attributes  `json:"attributes"`
	Subsection []interface{} `json:"subsections"`
}

type LinkMetadata struct {
	Url        string       `json:"url"`
	Attributes []Attributes `json:"attributes"`
}

type AccessionMetadata struct {
	Accno      string          `json:"accno"`
	Attributes []Attributes    `json:"attributes"`
	Sections   SectionMetadata `json:"section"`
}
