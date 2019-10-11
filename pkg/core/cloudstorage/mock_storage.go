package cloudstorage

import (
	"net/http"
	"net/url"
)

type MockStorage struct {
	PutRequest *http.Request
	GetURL     *url.URL
}

var _ Storage = &MockStorage{}

func (s *MockStorage) PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error) {
	return s.PutRequest, nil
}

func (s *MockStorage) PresignGetObject(name string) (*url.URL, error) {
	return s.GetURL, nil
}

func (s *MockStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.GetURL, nil
}

func (s *MockStorage) AccessType(header http.Header) AccessType {
	return AccessTypeDefault
}

func (s *MockStorage) StandardToProprietary(header http.Header) http.Header {
	return header
}

func (s *MockStorage) ProprietaryToStandard(header http.Header) http.Header {
	return header
}
