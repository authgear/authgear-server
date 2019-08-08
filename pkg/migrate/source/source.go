package source

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type Source interface {
	Download(sourceURL string) (localURL string, err error)
}

// Download the source from source url and return local file url (e.g. file://..)
// Currently only github url is supported, for the url that doesn't support
// function will return the original source url without altering
func Download(sourceURL string) (string, error) {
	s, err := getSource(sourceURL)
	if err != nil {
		return "", err
	}

	if s == nil {
		return sourceURL, nil
	}

	return s.Download(sourceURL)
}

func ClearCache() error {
	cacheDir, err := getTempCacheDirPath()
	if err != nil {
		return fmt.Errorf("unable to create tmp folder: %s", err.Error())
	}
	return os.RemoveAll(cacheDir)
}

func getSource(sourceURL string) (Source, error) {
	u, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid source url: %s", err.Error())
	}

	cacheDir, err := getTempCacheDirPath()
	if err != nil {
		return nil, fmt.Errorf("unable to create tmp folder: %s", err.Error())
	}

	var s Source
	switch u.Scheme {
	case "github":
		s = &Github{
			CacheDir: cacheDir,
		}
	}

	return s, nil
}

func getTempCacheDirPath() (string, error) {
	// ensure tmp folder
	dir := filepath.Join(os.TempDir(), "skygear-migrate-src")
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return "", err
	}
	return dir, nil
}
