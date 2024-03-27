package elasticsearch

type ReindexRequest struct {
	UserID string `json:"user_id"`
}

type ReindexResult struct {
	UserID       string `json:"user_id"`
	IsSuccess    bool   `json:"is_success"`
	ErrorMessage string `json:"error_message,omitempty"`
}
