package source

import (
	"fmt"
	"io"
	"io/ioutil"
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
	Owner    string
	Repo     string
	Path     string
	Ref      string
	Username string
	Password string
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
			), c.Username, c.Password); err != nil {
			return "", fmt.Errorf("unable to download: %s", err.Error())
		}

		tar := archiver.NewTarGz()
		// create top level folder when the same name as the tar file
		tar.ImplicitTopLevelFolder = true
		if err = tar.Unarchive(sourceCodeZipPath, g.CacheDir); err != nil {
			return "", fmt.Errorf("unable to unzip source: %s", err.Error())
		}
	}

	// the folder name of tarball are different for private and public repo
	// we unarchive the tarball into a folder and get the only folder name from
	// that folder
	sourceCodeDirName, err := getSrcFolderName(sourceCodeDirPath)
	if err != nil {
		return "", err
	}

	migrateSrcDirPath := filepath.Join(sourceCodeDirPath, sourceCodeDirName, c.Path)
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
	c.Username = u.User.Username()
	c.Password, _ = u.User.Password()

	return c, nil
}

func (g *Github) getSourceCodeDirPath(tmpDir string, c *config) string {
	return filepath.Join(
		tmpDir,
		fmt.Sprintf("%s-%s-%s", c.Owner, c.Repo, c.Ref),
	)
}

// getSrcFolderName searches and return the folder name inside the destination dir
// destination dir should only contain one folder
func getSrcFolderName(destinationDir string) (string, error) {
	var paths []string
	files, err := ioutil.ReadDir(destinationDir)
	if err != nil {
		return "", fmt.Errorf("unable to find source folder: %v", err.Error())
	}

	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, file.Name())
		}
	}

	if len(paths) != 1 {
		return "", fmt.Errorf("unable to find source: unarchive src should only have 1 folder, %d was find", len(paths))
	}

	return paths[0], err
}

func downloadFile(filepath string, url string, username string, password string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// This one line implements the authentication required for the task.
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server response: %d", resp.StatusCode)
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
