package hook

type Payload struct {
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
	Context map[string]interface{} `json:"context"`
}
