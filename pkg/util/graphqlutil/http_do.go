package graphqlutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/graphql-go/graphql"
)

// DoParams is the simplfied version of graphql.Params.
type DoParams struct {
	OperationName string                 `json:"operationName,omitempty"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

// HTTPDo is the HTTP version of graphql.Do.
func HTTPDo(r *http.Request, params DoParams) (result *graphql.Result, err error) {
	if params.Variables == nil {
		params.Variables = make(map[string]interface{})
	}

	requestBody, err := json.Marshal(params)
	if err != nil {
		return
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Length", strconv.Itoa(len(requestBody)))
	r.Body = ioutil.NopCloser(bytes.NewReader(requestBody))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		err = fmt.Errorf("unexpected status code %v: %v", resp.StatusCode, string(body))
		return
	}

	decoder := json.NewDecoder(resp.Body)
	result = &graphql.Result{}
	err = decoder.Decode(result)
	if err != nil {
		return
	}

	return
}
