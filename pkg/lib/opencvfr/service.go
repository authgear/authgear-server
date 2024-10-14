package opencvfr

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package opencvfr_test

type PersonService interface {
	Create(reqBody *openapi.CreatePersonSchema) (p *openapi.PersonSchema, err error)
	Get(id string) (p *openapi.PersonSchema, err error)
	Delete(id string) (err error)
	Update(reqBody *openapi.UpdatePersonSchema) (p *openapi.PersonSchema, err error)
	List(params *openapi.ListPersonsQuery) (p *openapi.ListPersonsSchema, err error)
	ListByCollection(collectionID string, params *openapi.ListPersonsQuery) (p *openapi.ListPersonsSchema, err error)
}

type CollectionService interface {
	Create(reqBody *openapi.CreateCollectionSchema) (c *openapi.CollectionSchema, err error)
	Get(id string) (c *openapi.CollectionSchema, err error)
	Delete(id string) (err error)
	Update(reqBody *openapi.UpdateCollectionSchema) (c *openapi.CollectionSchema, err error)
	LinkPerson(reqBody *openapi.LinkSchema) (l *openapi.LinkSchema, err error)
	UnlinkPerson(reqBody *openapi.LinkSchema) (err error)
}

type SearchService interface {
	Verify(reqBody *openapi.VerifyPersonSchema) (r *openapi.NullableVerifyPersonResultSchema, err error)
	Search(reqBody *openapi.SearchPersonSchema) (r []*openapi.SearchPersonResultSchema, err error)
	SearchLiveFace(reqBody *openapi.SearchLiveFaceScheme) (r *openapi.NullableSearchLivePersonResultSchema, err error)
}

type LivenessService interface {
	Check(reqBody *openapi.LivenessSchema) (r *openapi.LivenessResultSchema, err error)
}

type Service struct {
	Clock               clock.Clock
	AppID               config.AppID
	AuthenticatorConfig *config.AuthenticationConfig
	Person              PersonService
	Collection          CollectionService
	Search              SearchService
	Liveness            LivenessService
}

type VerifyFaceOption struct {
	OS openapi.OSEnum
}

// VerifyFace verifies if a face matches the given person and collection.
// If match, return nil.
// Otherwise, return error.
func (s *Service) VerifyFace(personID string, b64FaceImage string, opts *VerifyFaceOption) error {
	var os *openapi.OSEnum
	if opts != nil && opts.OS != "" {
		os = &opts.OS
	}
	// TODO (identity-week-demo): construct a mapping in db, appID <-> collectionID
	// collectionID := s.Store.GetCollectionID(appID)
	collectionID := "12edc48f-4b43-4240-90dd-213f3008932c"
	r, err := s.Search.SearchLiveFace(&openapi.SearchLiveFaceScheme{
		Os:           *openapi.NewNullableOSEnum(os),
		CollectionId: *openapi.NewNullableString(&collectionID),
		Image:        b64FaceImage,
	})

	if err != nil {
		return err
	}

	if r == nil {
		return ErrFaceNotFound
	}

	livePersonResult := r.Get()
	if livePersonResult == nil {
		return ErrFaceNotFound
	}

	// TODO (identity-week-demo): specify min-liveness somewhere
	const minLivenessScore float32 = 0.00001
	if livePersonResult.LivenessScore < minLivenessScore {
		return ErrFaceLivenessLow
	}

	var matched bool
	for _, p := range livePersonResult.Persons {
		if p.Id == personID {
			matched = true
			break
		}
	}

	if !matched {
		return ErrFaceNotMatch
	}

	// TODO (identity-week-demo): Handle opencvfr error codes, like
	// (ERR_INVALID_FACE_FOR_LIVENESS): Image is invalid for checking livenes
	// (ERR_FACE_EDGES_NOT_VISIBLE): The face is not in the center

	return nil
}
