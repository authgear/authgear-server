package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	ourAsset "github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

// used to clean file path
var sanitizedPathRe = regexp.MustCompile(`\A[/.]+`)

func clean(p string) string {
	sanitized := strings.Replace(sanitizedPathRe.ReplaceAllString(path.Clean(p), ""), "..", "", -1)
	// refs #426: S3 Asset Store is not able to put filename with `+` correctly
	sanitized = strings.Replace(sanitized, "+", "", -1)
	return sanitized
}

func AssetGetURLHandler(payload *router.Payload, response *router.Response) {
	payload.Req.ParseForm()

	// check whether the request is expired

	expiredAtUnix, err := strconv.ParseInt(payload.Req.Form.Get("expiredAt"), 10, 64)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect expiredAt to be an integer")
		return
	}
	expiredAt := time.Unix(expiredAtUnix, 0)
	if timeNow().After(expiredAt) {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "Access denied")
		return
	}

	// check the signature of the URL

	fileName := clean(payload.Params[0])
	signature := payload.Req.Form.Get("signature")

	signatureParser := payload.AssetStore.(ourAsset.SignatureParser)
	valid, err := signatureParser.ParseSignature(signature, fileName, expiredAt)
	if err != nil {
		log.Errorf("Failed to parse signature: %v", err)

		response.Err = skyerr.NewError(skyerr.PermissionDenied, "Access denied")
		return
	}

	if !valid {
		response.Err = skyerr.NewError(skyerr.InvalidSignature, "Invalid signature")
		return
	}

	// everything's right, proceed with the request

	conn := payload.DBConn
	asset := skydb.Asset{}
	if err := conn.GetAsset(fileName, &asset); err != nil {
		log.Errorf("Failed to get asset: %v", err)

		response.Err = skyerr.NewResourceFetchFailureErr("asset", fileName)
		return
	}

	response.Header().Set("Content-Type", asset.ContentType)
	response.Header().Set("Content-Length", strconv.FormatInt(asset.Size, 10))

	store := payload.AssetStore
	reader, err := store.GetFileReader(fileName)
	if err != nil {
		log.Errorf("Failed to get file reader: %v", err)

		response.Err = skyerr.NewResourceFetchFailureErr("asset", fileName)
		return
	}
	defer reader.Close()

	if _, err := io.Copy(response, reader); err != nil {
		// there is nothing we can do if error occurred after started
		// writing a response. Log.
		log.Errorf("Error writing file to response: %v", err)
	}
}

// AssetUploadURLHandler receives and persists a file to be associated by Record.
//
// Example curl:
//	curl -XPUT \
//		-H 'X-Skygear-API-Key: apiKey' \
//		-H 'Content-Type: text/plain' \
//		--data-binary '@file.txt' \
//		http://localhost:3000/files/filename
func AssetUploadURLHandler(payload *router.Payload, response *router.Response) {
	var (
		fileName, contentType string
	)

	fileName = clean(payload.Params[0])

	dir, file := filepath.Split(fileName)
	file = fmt.Sprintf("%s-%s", uuidNew(), file)

	fileName = filepath.Join(dir, file)
	contentType = payload.Req.Header.Get("Content-Type")

	if contentType == "" {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "Content-Type cannot be empty")
		return
	}

	written, tempFile, err := copyToTempFile(payload.Req.Body)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if written == 0 {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "Zero-byte content")
		return
	}

	assetStore := payload.AssetStore
	if err := assetStore.PutFileReader(fileName, tempFile, written, contentType); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}

	asset := skydb.Asset{
		Name:        fileName,
		ContentType: contentType,
		Size:        written,
	}

	conn := payload.DBConn
	if err := conn.SaveAsset(&asset); err != nil {
		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("asset", asset.Name)
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
