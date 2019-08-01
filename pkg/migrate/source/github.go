package source

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

type Github struct {
	CacheDir string
}

type config struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

func (g *Github) Download(sourceURL string) (string, error) {
	c, err := g.parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("invalid github source url: %s", err.Error())
	}

	sourceCodeDirPath := g.getSourceCodeDirPath(g.CacheDir, c)
	sourceCodeZipPath := fmt.Sprintf("%s.tar.gz", sourceCodeDirPath)
	if _, e := os.Stat(sourceCodeDirPath); os.IsNotExist(e) {
		// download and unzip source
		if err = downloadFile(
			sourceCodeZipPath,
			fmt.Sprintf(
				"https://api.github.com/repos/%s/%s/tarball/%s",
				c.Owner,
				c.Repo,
				c.Ref,
			)); err != nil {
			return "", fmt.Errorf("unable to download: %s", err.Error())
		}

		if err = archiver.Unarchive(sourceCodeZipPath, g.CacheDir); err != nil {
			return "", fmt.Errorf("unable to unzip source: %s", err.Error())
		}
	}

	migrateSrcDirPath := filepath.Join(sourceCodeDirPath, c.Path)
	if _, e := os.Stat(migrateSrcDirPath); os.IsNotExist(e) {
		return "", fmt.Errorf("unable to find source: %v: no such directory", c.Path)
	}

	return fmt.Sprintf("file://%s", migrateSrcDirPath), nil
}

func (g *Github) parse(sourceURL string) (*config, error) {
	c := &config{}
	u, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid source url: %s", err.Error())
	}

	c.Owner = u.Host
	p := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(p) < 1 {
		return nil, fmt.Errorf("invalid source url: missing host")
	}
	c.Repo = p[0]
	c.Path = strings.Join(p[1:], "/")
	c.Ref = u.Fragment
	return c, nil
}

func (g *Github) getSourceCodeDirPath(tmpDir string, c *config) string {
	return filepath.Join(
		tmpDir,
		fmt.Sprintf("%s-%s-%s", c.Owner, c.Repo, c.Ref),
	)
}

func downloadFile(filepath string, url string) error {
	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
