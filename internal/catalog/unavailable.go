package catalog

import (
	"context"
	"errors"
)

// ErrBackendUnavailable indicates that no Ballerina Central documentation
// backend is wired.
var ErrBackendUnavailable = errors.New("Ballerina Central documentation backend is not configured")

// UnavailableDocumentationService is used until the documentation backend is
// implemented.
type UnavailableDocumentationService struct{}

// NewUnavailableDocumentationService returns a documentation service that
// reports that its backend has not been configured.
func NewUnavailableDocumentationService() DocumentationService {
	return UnavailableDocumentationService{}
}

// Documentation reports that the backend has not been configured.
func (UnavailableDocumentationService) Documentation(context.Context, PackageSelector) (PackageDocumentation, error) {
	return PackageDocumentation{}, ErrBackendUnavailable
}
