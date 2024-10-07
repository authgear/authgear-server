package openapi

// Note this response is not documented by openapi.json, but instead found in python sdk

type APIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
