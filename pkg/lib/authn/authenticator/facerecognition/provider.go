package facerecognition

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/opencvfr"
	opencvfrapi "github.com/authgear/authgear-server/pkg/lib/opencvfr/api"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type OpenCVFRService interface {
	CreatePerson(opts *opencvfr.CreatePersonOptions) (p *opencvfr.CreatePersonOutput, err error)
	VerifyLiveFace(opts *opencvfr.VerifyLiveFaceOption) error
}

type Provider struct {
	Store    *Store
	OpenCVFR OpenCVFRService
	Clock    clock.Clock
}

func (p *Provider) Get(userID string, id string) (*authenticator.FaceRecognition, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.FaceRecognition, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.FaceRecognition) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*authenticator.FaceRecognition, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(id string, userID string, frSpec *authenticator.FaceRecognitionSpec, isDefault bool, kind string) (*authenticator.FaceRecognition, error) {
	if id == "" {
		id = uuid.New()
	}

	// TODO (identity-week-demo): Check if person face already exists in opencvfr database,  in other projects (collections) first, which is possible
	person, err := p.OpenCVFR.CreatePerson(&opencvfr.CreatePersonOptions{
		Name: "authgear-" + userID,
		B64ImageList: []string{
			frSpec.B64ImageString,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create new face_recognition authenticator for user (%s): %w", userID, err)
	}

	a := &authenticator.FaceRecognition{
		ID:               id,
		UserID:           userID,
		OpenCVFRPersonID: person.OpenCVFRPersonID,
		IsDefault:        isDefault,
		Kind:             kind,
	}
	return a, nil
}

func (p *Provider) Create(a *authenticator.FaceRecognition) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Authenticate(a *authenticator.FaceRecognition, b64Image string) error {
	return p.OpenCVFR.VerifyLiveFace(&opencvfr.VerifyLiveFaceOption{
		PersonID:     a.OpenCVFRPersonID,
		B64FaceImage: b64Image,
	})
}

func (p *Provider) ParseError(apiErr *opencvfr.APIError) error {
	switch apiErr.Name {
	case opencvfr.NoMatchingFaceFound:
		return NoMatchingFaceFound.New(apiErr.Message)
	case opencvfr.SpoofedImageDetected:
		return SpoofedImageDetected.New(apiErr.Message)
	case opencvfr.InvalidInput:
		if apiErr.RawErr == nil {
			return LowImageQuality.New(apiErr.Message)
		} else {
			switch apiErr.RawErr.ErrCode() {
			case opencvfrapi.FaceTooSmall:
				return FaceTooSmall.New(apiErr.Message)
			case opencvfrapi.NoFacesFound:
				return NoFaceDetected.New(apiErr.Message)
			case opencvfrapi.FaceRotation:
				return FaceRotated.New(apiErr.Message)
			case opencvfrapi.FaceEdgesNotVisible:
				return FaceEdgesNotVisible.New(apiErr.Message)
			case opencvfrapi.FaceOccluded:
				return FaceCovered.New(apiErr.Message)
			case opencvfrapi.FaceTooClose:
				return FaceTooClose.New(apiErr.Message)
			case opencvfrapi.FaceCropped:
				return FaceCropped.New(apiErr.Message)
			case opencvfrapi.InvalidFaceForLiveness:
				return LowImageQuality.New(apiErr.Message)
			case opencvfrapi.MultipleFaces:
				return MultipleFaces.New(apiErr.Message)
			case opencvfrapi.BlurryImage:
				return LowImageQuality.New(apiErr.Message)
			default:
				return LowImageQuality.New(apiErr.Message)
			}
		}
	case opencvfr.ServiceUnavailable:
		fallthrough
	case opencvfr.OperationNotAllowed:
		fallthrough
	case opencvfr.UnexpectedError:
		fallthrough
	default:
		return apiErr
	}
}

func sortAuthenticators(as []*authenticator.FaceRecognition) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
