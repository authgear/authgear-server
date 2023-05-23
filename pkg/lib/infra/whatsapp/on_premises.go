package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WhatsappOnPremisesClient struct {
	HTTPClient *http.Client
	Endpoint   *url.URL
}

func NewWhatsappOnPremisesClient(
	config *config.WhatsappConfig) *WhatsappOnPremisesClient {
	if !config.Enabled || config.APIEndpoint == "" {
		return nil
	}
	endpoint, err := url.Parse(config.APIEndpoint)
	if err != nil {
		panic(err)
	}
	return &WhatsappOnPremisesClient{
		HTTPClient: http.DefaultClient,
		Endpoint:   endpoint,
	}
}

func (c *WhatsappOnPremisesClient) SendTemplate(
	authToken string,
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
	namespace string) error {
	body := &SendTemplateRequest{
		RecipientType: "individual",
		To:            to,
		Type:          "template",
		Template: &Template{
			Name: templateName,
			Language: &TemplateLanguage{
				Policy: "deterministic",
				Code:   templateLanguage,
			},
			Components: templateComponents,
			Namespace:  &namespace,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := c.Endpoint.JoinPath("/v1/messages")

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrUnexpectedStatus
	}

	return nil
}

func (c *WhatsappOnPremisesClient) Login(
	username string,
	password string,
) (*LoginResponseUser, error) {
	url := c.Endpoint.JoinPath("/v1/users/login")
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrUnexpectedStatus
	}

	loginHTTPResponseBytes, err := io.ReadAll(resp.Body)

	var loginResponse LoginResponse
	err = json.Unmarshal(loginHTTPResponseBytes, &loginResponse)
	if err != nil {
		return nil, err
	}

	if len(loginResponse.Users) < 1 {
		return nil, ErrUnexpectedLoginResponse
	}

	return &loginResponse.Users[0], nil
}
