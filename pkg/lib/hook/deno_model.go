package hook

type RunRequest struct {
	Script string      `json:"script"`
	Input  interface{} `json:"input"`
}

type Stream struct {
	String    string `json:"string,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

type ErrorCode string

const (
	ErrorCodeRunTimout ErrorCode = "run_timeout"
	ErrorCodeUnknown   ErrorCode = "unknown"
)

type RunResponse struct {
	Error     string      `json:"error,omitempty"`
	ErrorCode ErrorCode   `json:"error_code,omitempty"`
	Output    interface{} `json:"output,omitempty"`
	Stderr    *Stream     `json:"stderr,omitempty"`
	Stdout    *Stream     `json:"stdout,omitempty"`
}

type CheckRequest struct {
	Script string `json:"script"`
}

type CheckResponse struct {
	Stderr string `json:"stderr,omitempty"`
}
