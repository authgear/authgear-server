package hook

type RunRequest struct {
	Script string      `json:"script"`
	Input  interface{} `json:"input"`
}

type Stream struct {
	String    string `json:"string,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

type RunResponse struct {
	Error  string      `json:"error,omitempty"`
	Output interface{} `json:"output,omitempty"`
	Stderr *Stream     `json:"stderr,omitempty"`
	Stdout *Stream     `json:"stdout,omitempty"`
}
