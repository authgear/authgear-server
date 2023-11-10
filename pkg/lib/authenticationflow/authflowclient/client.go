package authflowclient

import (
	"net/http"
)

type Client struct{}

func (c *Client) Create(flowReference FlowReference, urlQuery string) (*FlowResponse, error) {
	panic("not yet implemented")
}

func (c *Client) Get(stateToken string) (*FlowResponse, error) {
	panic("not yet implemented")
}

func (c *Client) Input(w http.ResponseWriter, stateToken string, input map[string]interface{}) (*FlowResponse, error) {
	panic("not yet implemented")
}
