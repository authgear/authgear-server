package api

import (
	"fmt"
)

// NOTE(DEV-2175): There are no official doc for complete list of error codes. Instead, it exists as an excel file
type ErrCode string

const (
	// common
	MaintenanceMode       ErrCode = "ERR_MAINTENANCE_MODE"
	RateLimitExceeded     ErrCode = "ERR_RATE_LIMIT_EXCEEDED"
	InvalidAPIKey         ErrCode = "ERR_INVALID_API_KEY"
	InvalidToken          ErrCode = "ERR_INVALID_TOKEN"
	ExpiredSubscription   ErrCode = "ERR_EXPIRED_SUBSCRIPTION"
	MasterKeyCannotBeUsed ErrCode = "ERR_MASTER_KEY_CANNOT_BE_USED"
	DisabledAPIKey        ErrCode = "ERR_DISABLED_API_KEY"
	RestrictedAPI         ErrCode = "ERR_RESTRICTED_API"

	// liveness, search, compare, person
	FaceTooSmall           ErrCode = "ERR_FACE_TOO_SMALL"
	ImageCouldNotBeDecoded ErrCode = "ERR_IMAGE_COULD_NOT_BE_DECODED"
	NoFacesFound           ErrCode = "ERR_NO_FACES_FOUND"
	FaceRotation           ErrCode = "ERR_FACE_ROTATION"
	FaceEdgesNotVisible    ErrCode = "ERR_FACE_EDGES_NOT_VISIBLE"
	FaceOccluded           ErrCode = "ERR_FACE_OCCLUDED"
	FaceTooClose           ErrCode = "ERR_FACE_TOO_CLOSE"
	FaceCropped            ErrCode = "ERR_FACE_CROPPED"
	InvalidFaceForLiveness ErrCode = "ERR_INVALID_FACE_FOR_LIVENESS"
	MultipleFaces          ErrCode = "ERR_MULTIPLE_FACES"
	BlurryImage            ErrCode = "ERR_BLURRY_IMAGE"

	// search, get/update/delete collection, get/update/delete person
	EntityNotFound ErrCode = "ERR_ENTITY_NOT_FOUND"

	// create/update/delete collection, create/update/delete person
	DuplicateEntity ErrCode = "ERR_DUPLICATE_ENTITY"
	ReadOnlyAPIKey  ErrCode = "ERR_READ_ONLY_API_KEY"

	// create person
	PersonLimitExceeded ErrCode = "ERR_PERSON_LIMIT_EXCEEDED"

	// add/remove person images
	UpdateResultsInTooManyThumbnails ErrCode = "ERR_UPDATE_RESULTS_IN_TOO_MANY_THUMBNAILS"

	// update person images
	UpdateResultsInNoThumbnail ErrCode = "ERR_UPDATE_RESULTS_IN_NO_THUMBNAILS"

	// delete person images
	DeleteResultsInNoThumbnail ErrCode = "ERR_DELETE_RESULTS_IN_NO_THUMBNAILS"

	// crop
	NotStandardCrop ErrCode = "ERR_NOT_STANDARD_CROP"
)

type OpenCVFRAPIError struct {
	HTTPStatusCode  int    `json:"httpStatusCode"`
	Message         string `json:"message"`
	OpenCVFRErrCode string `json:"openCVFRErrCode"`
	RetryAfter      int    `json:"retryAfter"`
}

func (e *OpenCVFRAPIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("opencvfr api error (%s): %s", e.OpenCVFRErrCode, e.Message)
}

// ErrCode returns pre-defined ErrCode if matched. Otherwise, return empty string
func (e *OpenCVFRAPIError) ErrCode() ErrCode {
	ec := ErrCode(e.OpenCVFRErrCode)
	switch ec {
	case MaintenanceMode:
		return ec
	case RateLimitExceeded:
		return ec
	case InvalidAPIKey:
		return ec
	case InvalidToken:
		return ec
	case ExpiredSubscription:
		return ec
	case MasterKeyCannotBeUsed:
		return ec
	case DisabledAPIKey:
		return ec
	case RestrictedAPI:
		return ec
	case FaceTooSmall:
		return ec
	case ImageCouldNotBeDecoded:
		return ec
	case NoFacesFound:
		return ec
	case FaceRotation:
		return ec
	case FaceEdgesNotVisible:
		return ec
	case FaceOccluded:
		return ec
	case FaceTooClose:
		return ec
	case FaceCropped:
		return ec
	case InvalidFaceForLiveness:
		return ec
	case MultipleFaces:
		return ec
	case BlurryImage:
		return ec
	case EntityNotFound:
		return ec
	case DuplicateEntity:
		return ec
	case ReadOnlyAPIKey:
		return ec
	case PersonLimitExceeded:
		return ec
	case UpdateResultsInTooManyThumbnails:
		return ec
	case UpdateResultsInNoThumbnail:
		return ec
	case DeleteResultsInNoThumbnail:
		return ec
	case NotStandardCrop:
		return ec
	default:
		return ""
	}
}

type OpenCVFRValidationError struct {
	Details string `json:"details"`
}

func (e *OpenCVFRValidationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("opencvfr validation error: %s", e.Details)
}
