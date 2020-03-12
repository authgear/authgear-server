package model

type AssignmentType string

const (
	AssignmentTypeAuth          AssignmentType = "auth"
	AssignmentTypeAsset         AssignmentType = "asset"
	AssignmentTypeMicroservices AssignmentType = "microservices"
)

type Domain struct {
	ID         string         `db:"id"`
	AppID      string         `db:"app_id"`
	Domain     string         `db:"domain"`
	Assignment AssignmentType `db:"assignment"`
}
