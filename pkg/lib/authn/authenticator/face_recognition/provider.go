package face_recognition

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/opencvfr"
	opencvfropenapi "github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type OpenCVFRService interface {
	VerifyFace(appID string, personID string, b64FaceImage string, opts *opencvfr.VerifyFaceOption) error
}

type Provider struct {
	AppID    config.AppID
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

	a := &authenticator.FaceRecognition{
		ID:               id,
		UserID:           userID,
		OpenCVFRPersonID: frSpec.OpenCVFRPersonID,
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
	return p.OpenCVFR.VerifyFace(string(p.AppID), a.OpenCVFRPersonID, b64Image, &opencvfr.VerifyFaceOption{
		OS: opencvfropenapi.DESKTOP, // TODO: check user os
	})
}

func sortAuthenticators(as []*authenticator.FaceRecognition) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
