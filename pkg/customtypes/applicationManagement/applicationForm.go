package applicationManagement

import "time"

type FormOption struct {
	Label string
	Value string
}

type FormField struct {
	Name    string
	Type    string
	Options []FormOption
}

type FormQuestion struct {
	Label    string
	Required bool
	Fields   []FormField
}

type ApplicationForm struct {
	Questions          []FormQuestion
	HasHiddenQuestions bool
}

type PreparedAnswer struct {
	Label     string   `json:"label"`
	FieldName string   `json:"fieldName"`
	Value     string   `json:"value"`
	Values    []string `json:"values,omitempty"`
}

type MissingAnswer struct {
	Label     string `json:"label"`
	FieldName string `json:"fieldName"`
}

type Submission struct {
	FirstName  string
	LastName   string
	Email      string
	Answers    []PreparedAnswer
	Resume     []byte
	ResumeName string
	ResumeType string
}

type ExecutionLogEntry struct {
	At     time.Time `json:"at"`
	Step   string    `json:"step"`
	Detail string    `json:"detail,omitempty"`
}
