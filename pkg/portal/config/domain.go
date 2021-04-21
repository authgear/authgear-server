package config

type DomainImplementationType string

const (
	DomainImplementationTypeNone       DomainImplementationType = ""
	DomainImplementationTypeKubernetes DomainImplementationType = "kubernetes"
)
