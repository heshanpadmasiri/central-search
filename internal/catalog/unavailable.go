package catalog

import (
	"context"
	"errors"
)

// ErrBackendUnavailable indicates that no Ballerina Central backend is wired.
var ErrBackendUnavailable = errors.New("Ballerina Central backend is not configured")

// UnavailableService is used until the Central backend is implemented.
type UnavailableService struct{}

// NewUnavailableService returns a catalog service that reports that the
// backend has not been configured.
func NewUnavailableService() Service {
	return UnavailableService{}
}

// Search reports that the backend has not been configured.
func (UnavailableService) Search(context.Context, string) ([]PackageSummary, error) {
	return nil, ErrBackendUnavailable
}

// Documentation reports that the backend has not been configured.
func (UnavailableService) Documentation(context.Context, PackageSelector) (PackageDocumentation, error) {
	return PackageDocumentation{}, ErrBackendUnavailable
}
