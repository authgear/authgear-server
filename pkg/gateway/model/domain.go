package model

type AssignmentType string

const (
	AssignmentTypeAuth          AssignmentType = "auth"
	AssignmentTypeAsset         AssignmentType = "asset"
	AssignmentTypeMicroservices AssignmentType = "microservices"

	// AssignmentTypeDefault is used for app default domain only
	// Default domain supports gear subdomains
	// e.g. If an app has default domain `app1.skygearapp.com`, gear subdomains
	// `accounts.app1.skygearapp.com` and `asset.app1.skygearapp.com` will be
	// support automatically
	// Custom domain doesn't support default assignment
	AssignmentTypeDefault AssignmentType = "default"
)

type Domain struct {
	ID         string         `db:"id"`
	AppID      string         `db:"app_id"`
	Domain     string         `db:"domain"`
	Assignment AssignmentType `db:"assignment"`
}
