package store

type ReferenceValidator interface {

	Validate(resourceType, resourceID string) bool

}