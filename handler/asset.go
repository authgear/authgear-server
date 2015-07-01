package handler

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
	"github.com/oursky/ourd/uuid"
)

func AssetGetURLHandler(payload *router.Payload, response *router.Response) {
	fileName := payload.Params[0]

	conn := payload.DBConn
	asset := oddb.Asset{}
	if err := conn.GetAsset(fileName, &asset); err != nil {
		log.Errorf("Failed to get asset: %v", err)

		response.Err = oderr.NewResourceFetchFailureErr("asset", fileName)
		return
	}

	response.Header().Set("Content-Type", asset.ContentType)

	store := payload.AssetStore
	reader, err := store.GetFileReader(fileName)
	if err != nil {
		log.Errorf("Failed to get file reader: %v", err)

		response.Err = oderr.NewResourceFetchFailureErr("asset", fileName)
		return
	}
	defer reader.Close()

	if _, err := io.Copy(response, reader); err != nil {
		// there is nothing we can do if error occurred after started
		// writing a response. Log.
		log.Fatalf("Error writing file to response: %v", err)
	}
}

// AssetUploadURLHandler receives and persists a file to be associated by Record.
//
// Example curl:
//	curl -XPUT \
//		-H 'X-Ourd-API-Key: apiKey' \
//		-H 'Content-Type: text/plain' \
//		--data-binary '@file.txt' \
//		http://localhost:3000/files/filename
func AssetUploadURLHandler(payload *router.Payload, response *router.Response) {
	var (
		fileName, contentType string
	)

	dir, file := filepath.Split(payload.Params[0])
	file = fmt.Sprintf("%s-%s", uuid.New(), file)

	fileName = filepath.Join(dir, file)
	contentType = payload.Req.Header.Get("Content-Type")

	if contentType == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("Content-Type cannot be empty"))
		return
	}

	written, tempFile, err := copyToTempFile(payload.Req.Body)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if written == 0 {
		response.Err = oderr.NewRequestInvalidErr(errors.New("Zero-byte content"))
	}

	assetStore := payload.AssetStore
	if err := assetStore.PutFileReader(fileName, tempFile, written, contentType); err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}

	asset := oddb.Asset{
		Name:        fileName,
		ContentType: contentType,
	}

	conn := payload.DBConn
	if err := conn.SaveAsset(&asset); err != nil {
		response.Err = oderr.NewResourceSaveFailureErrWithStringID("asset", asset.Name)
		return
	}

	response.Result = struct {
		Type string `json:"$type"`
		Name string `json:"$name"`
	}{"asset", asset.Name}
}

func copyToTempFile(src io.Reader) (written int64, tempFile *os.File, err error) {
	tempFile, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}
	written, err = io.Copy(tempFile, src)
	if err != nil {
		cleanupFile(tempFile)
		tempFile = nil
		return
	}
	if _, err = tempFile.Seek(0, 0); err != nil {
		cleanupFile(tempFile)
		tempFile = nil
		return
	}
	return
}

func cleanupFile(f *os.File) error {
	closeErr := f.Close()
	if closeErr != nil {
		log.Errorf("Failed to close tempFile %s: %v", f.Name(), closeErr)
		return closeErr
	}

	if err := os.Remove(f.Name()); err != nil {
		log.Errorf("Failed to remove file %s: %v", f.Name(), err)
		return err
	}

	return nil
}
