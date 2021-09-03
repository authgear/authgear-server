package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetOAuth2Config(credentialsJSONFilePath string, scope ...string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(credentialsJSONFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, scope...)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	return config, nil
}

func GetTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}
	return token, nil
}

func GetTokenFromFile(path string) (*oauth2.Token, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read token file: %v", err)
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(b, &token)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode token file: %v", err)
	}

	return token, nil
}

func SaveToken(path string, token *oauth2.Token) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Unable to encode token to json: %v", err)
	}

	err = ioutil.WriteFile(path, tokenJSON, 0600)
	if err != nil {
		return fmt.Errorf("Unable to save token.json: %v", err)
	}

	return nil
}

func GetGoogleSheetsService(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*sheets.Service, error) {
	client := config.Client(ctx, token)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}

	return srv, nil
}
