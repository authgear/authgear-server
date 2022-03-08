package imageproxy

import (
	"fmt"
	"net/http"
	"strings"
)

//go:generate mockgen -source=imageproxy.go -destination=imageproxy_mock.go -package imageproxy

type ExtractKey func(r *http.Request) string

func getKey(extractKey ExtractKey, r *http.Request) string {
	key := extractKey(r)
	if strings.HasPrefix(key, "/") {
		return strings.TrimPrefix(key, "/")
	}
	return key
}

type Director interface {
	Director(r *http.Request)
}

type GCPGCSDirector struct {
	ExtractKey ExtractKey
	BucketName string
}

func (d GCPGCSDirector) Director(r *http.Request) {
	scheme := "https"
	host := "storage.googleapis.com"
	key := getKey(d.ExtractKey, r)
	path := fmt.Sprintf("/%s/%s", d.BucketName, key)

	r.URL.Scheme = scheme
	r.URL.Host = host
	r.URL.Path = path
	r.URL.RawQuery = ""
	r.URL.RawFragment = ""

	r.Host = host
}

var _ Director = GCPGCSDirector{}

type AWSS3Director struct {
	ExtractKey ExtractKey
	BucketName string
	Region     string
}

func (d AWSS3Director) Director(r *http.Request) {
	scheme := "https"
	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", d.BucketName, d.Region)
	key := getKey(d.ExtractKey, r)
	path := fmt.Sprintf("/%s", key)

	r.URL.Scheme = scheme
	r.URL.Host = host
	r.URL.Path = path
	r.URL.RawQuery = ""
	r.URL.RawFragment = ""

	r.Host = host
}

var _ Director = AWSS3Director{}

type AzureBlobStorageDirector struct {
	ExtractKey     ExtractKey
	StorageAccount string
	Container      string
}

var _ Director = AzureBlobStorageDirector{}

func (d AzureBlobStorageDirector) Director(r *http.Request) {
	scheme := "https"
	host := fmt.Sprintf("%s.blob.core.windows.net", d.StorageAccount)
	key := getKey(d.ExtractKey, r)
	path := fmt.Sprintf("/%s/%s", d.Container, key)

	r.URL.Scheme = scheme
	r.URL.Host = host
	r.URL.Path = path
	r.URL.RawQuery = ""
	r.URL.RawFragment = ""

	r.Host = host
}
