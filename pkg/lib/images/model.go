package images

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	QueryMetadata = "x-authgear-metadata"
)

type UploadedByType string

var FileMetaSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"user_id": { "type": "string" },
			"uploaded_by": {
				"type": "string",
				"enum": ["user", "admin_api"]
			}
		},
		"required": ["user_id", "uploaded_by"]
	}
`)

const (
	UploadedByTypeUser     UploadedByType = "user"
	UploadedByTypeAdminAPI UploadedByType = "admin_api"
)

type File struct {
	ID        string
	Size      int64
	CreatedAt time.Time
	Metadata  *FileMetadata
}

type FileMetadata struct {
	UserID     string         `json:"user_id,omitempty"`
	UploadedBy UploadedByType `json:"uploaded_by,omitempty"`
}

func EncodeFileMetaData(metadata *FileMetadata) (string, error) {
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

func DecodeFileMetadata(encoded string) (*FileMetadata, error) {
	if encoded == "" {
		return nil, apierrors.NewInvalid("missing metadata")
	}

	jsonBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	err = FileMetaSchema.Validator().ValidateWithMessage(
		bytes.NewReader(jsonBytes),
		"invalid file metadata",
	)
	if err != nil {
		return nil, err
	}

	metadata := &FileMetadata{}
	err = json.Unmarshal(jsonBytes, metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
