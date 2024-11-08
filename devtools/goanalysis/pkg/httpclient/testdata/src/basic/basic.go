package basic

import (
	"context"
	"net/http"
	aliasedhttp "net/http"
)

func ConstructingHTTPClient() {
	_ = http.Client{}         // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
	_ = &http.Client{}        // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
	_ = aliasedhttp.Client{}  // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
	_ = &aliasedhttp.Client{} // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
}

func UseHTTPDefaultClient() {
	_ = http.DefaultClient // want `Using http.DefaultClient is forbidden. Use httputil.NewExternalClient instead.`
}

func ConstructingHTTPRequest() {
	_, _ = aliasedhttp.NewRequestWithContext(context.Background(), "GET", "", nil)
	_, _ = http.NewRequestWithContext(context.Background(), "GET", "", nil)

	_, _ = http.NewRequest("GET", "", nil)        // want `Calling http.NewRequest is forbidden. Use http.NewRequestWithContext instead.`
	_, _ = aliasedhttp.NewRequest("GET", "", nil) // want `Calling http.NewRequest is forbidden. Use http.NewRequestWithContext instead.`
}

func UseHTTPDefaultClientImplicitly() {
	_, _ = http.Get("")           // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = http.Head("")          // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = http.Post("", "", nil) // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = http.PostForm("", nil) // want `context.Context is lost. Use http.Client.Do instead.`
}

func UseMethodsOtherThanDo() {
	nonPointerHTTPClient := http.Client{}         // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
	_, _ = nonPointerHTTPClient.Get("")           // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = nonPointerHTTPClient.Head("")          // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = nonPointerHTTPClient.Post("", "", nil) // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = nonPointerHTTPClient.PostForm("", nil) // want `context.Context is lost. Use http.Client.Do instead.`

	pointerHTTPClient := &http.Client{}        // want `Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.`
	_, _ = pointerHTTPClient.Get("")           // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = pointerHTTPClient.Head("")          // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = pointerHTTPClient.Post("", "", nil) // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = pointerHTTPClient.PostForm("", nil) // want `context.Context is lost. Use http.Client.Do instead.`

	type EmbeddedStructClient struct {
		http.Client
	}
	embeddedStructClient := EmbeddedStructClient{}
	_, _ = embeddedStructClient.Get("")           // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedStructClient.Head("")          // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedStructClient.Post("", "", nil) // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedStructClient.PostForm("", nil) // want `context.Context is lost. Use http.Client.Do instead.`

	type EmbeddedPointerClient struct {
		*http.Client
	}
	embeddedPointerClient := EmbeddedPointerClient{}
	_, _ = embeddedPointerClient.Get("")           // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedPointerClient.Head("")          // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedPointerClient.Post("", "", nil) // want `context.Context is lost. Use http.Client.Do instead.`
	_, _ = embeddedPointerClient.PostForm("", nil) // want `context.Context is lost. Use http.Client.Do instead.`
}
