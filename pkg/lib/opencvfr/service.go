package opencvfr

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slice"
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

type OpenCVFRCollectionIDMapStore interface {
	Get(appID string) (m *AuthgearAppIDOpenCVFRCollectionIDMap, err error)
	Create(m *AuthgearAppIDOpenCVFRCollectionIDMap) (err error)
}

type Service struct {
	Clock                        clock.Clock
	AppID                        config.AppID
	AuthenticatorConfig          *config.AuthenticationConfig
	Person                       PersonService
	Collection                   CollectionService
	Search                       SearchService
	Liveness                     LivenessService
	OpenCVFRCollectionIDMapStore OpenCVFRCollectionIDMapStore
}

type VerifyFaceOption struct {
	PersonID     string
	B64FaceImage string
}

// VerifyFace verifies if a face matches the given person and collection.
// If match, return nil.
// Otherwise, return error.
func (s *Service) VerifyFace(opts *VerifyFaceOption) error {
	collection, err := s.getCollection(string(s.AppID))
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection for app id (%s) not found", s.AppID)
	}
	ps, err := s.Search.Search(&openapi.SearchPersonSchema{
		CollectionId: *openapi.NewNullableString(&collection.Id),
		Images:       []string{opts.B64FaceImage},
		SearchMode:   *openapi.NewNullableSearchModeEnum(openapi.FAST.Ptr()),
	})

	if err != nil {
		return s.parseOpenCVFRError(err)
	}

	if len(ps) == 0 {
		return newNoMatchingFaceFoundError()
	}

	for _, p := range ps {
		if p == nil {
			continue
		}
		if p.GetId() == opts.PersonID {
			return nil
		}
	}
	return newNoMatchingFaceFoundError()
}

type VerifyLiveFaceOption struct {
	PersonID         string
	B64FaceImage     string
	MinLivenessScore float32
}

// VerifyLiveFace verifies if a face matches the given person and collection. It also check liveness of the image.
// If match, return nil.
// Otherwise, return error.
func (s *Service) VerifyLiveFace(opts *VerifyLiveFaceOption) error {
	collection, err := s.getCollection(string(s.AppID))
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection for app id (%s) not found", s.AppID)
	}
	r, err := s.Search.SearchLiveFace(&openapi.SearchLiveFaceScheme{
		CollectionId: *openapi.NewNullableString(&collection.Id),
		Image:        opts.B64FaceImage,
	})

	if err != nil {
		return s.parseOpenCVFRError(err)
	}

	if r == nil {
		return newNoMatchingFaceFoundError()
	}

	livePersonResult := r.Get()
	if livePersonResult == nil {
		return newNoMatchingFaceFoundError()
	}

	if livePersonResult.LivenessScore < opts.MinLivenessScore {
		return newSpoofedImageDetectedError()
	}

	var matched bool
	for _, p := range livePersonResult.Persons {
		if p.Id == opts.PersonID {
			matched = true
			break
		}
	}

	if !matched {
		return newNoMatchingFaceFoundError()
	}

	return nil
}

type CreatePersonOptions struct {
	Name         string
	B64ImageList []string
	DateOfBirth  *time.Time
	Nationality  string
	Notes        string
}

type CreatePersonOutput struct {
	OpenCVFRPersonID   string   `json:"opencv_fr_user_id,omitempty"`
	OpenCVFRPersonName string   `json:"opencv_fr_user_name,omitempty"`
	B64ImageList       []string `json:"b64_image_list,omitempty"`
}

func (s *Service) CreatePerson(opts *CreatePersonOptions) (p *CreatePersonOutput, err error) {
	collection, err := s.getCollection(string(s.AppID))
	if err != nil {
		return nil, err
	}

	if collection == nil {
		// no existing collection, create one
		c, err := s.createCollection(string(s.AppID))
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, fmt.Errorf("failed to create collection")
		}
		collection = c
	}
	schema := &openapi.CreatePersonSchema{
		Collections: []string{collection.GetId()},
	}
	if opts != nil {
		if opts.Name != "" {
			schema.Name = *openapi.NewNullableString(&opts.Name)
		}
		if len(opts.B64ImageList) > 0 {
			schema.Images = opts.B64ImageList
		}
		if opts.Notes != "" {
			schema.Notes = *openapi.NewNullableString(&opts.Notes)
		}
	}
	resp, err := s.Person.Create(schema)

	if err != nil {
		return nil, s.parseOpenCVFRError(err)
	}

	output := &CreatePersonOutput{
		OpenCVFRPersonID:   resp.Id,
		OpenCVFRPersonName: resp.GetName(),
		B64ImageList: slice.Map(
			resp.Thumbnails,
			func(t openapi.ThumbnailSchema) string { return t.GetThumbnail() },
		),
	}
	return output, nil
}

func (s *Service) getCollection(appID string) (c *openapi.CollectionSchema, err error) {
	m, err := s.OpenCVFRCollectionIDMapStore.Get(appID)
	if err != nil {
		err = s.parseOpenCVFRError(err)
		return nil, fmt.Errorf("failed to get collection id of app id (%s): %w", appID, err)
	}
	if m == nil {
		return nil, nil
	}

	// TODO (identity-week-demo): Handle mismatch of records
	c, err = s.Collection.Get(m.OpenCVFRCollectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection of id (%s): %w", appID, err)
	}
	if c == nil || c.GetId() == "" {
		return nil, nil
	}
	return c, nil
}

func (s *Service) createCollection(appID string) (c *openapi.CollectionSchema, err error) {
	des := "This collection is a one-to-one map towards authgear project '" + appID + "'"
	schema := &openapi.CreateCollectionSchema{
		Name:        "authgear-" + appID,
		Description: *openapi.NewNullableString(&des),
	}
	c, err = s.Collection.Create(schema)

	if err != nil {
		err = s.parseOpenCVFRError(err)
		return nil, err
	}

	// TODO (identity-week-demo): Make these two operations atomic

	err = s.OpenCVFRCollectionIDMapStore.Create(&AuthgearAppIDOpenCVFRCollectionIDMap{
		AppID:                appID,
		OpenCVFRCollectionID: c.GetId(),
	})

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) parseOpenCVFRError(err error) error {
	if apiErr := getAPIError(err); apiErr != nil {
		return apiErr
	}
	return err
}
