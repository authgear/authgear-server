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
}

type Service struct {
	Clock               clock.Clock
	AppID               config.AppID
	AuthenticatorConfig *config.AuthenticationConfig
	Person              PersonService
	Collection          CollectionService
}
