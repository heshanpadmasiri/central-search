package catalog

import (
	"errors"
	"testing"
)

func TestParsePackageSelector(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  PackageSelector
	}{
		{name: "unqualified", input: "http", want: PackageSelector{Package: "http"}},
		{name: "qualified", input: "ballerina/http", want: PackageSelector{Organization: "ballerina", Package: "http"}},
		{name: "trims surrounding whitespace", input: "  ballerina/http  ", want: PackageSelector{Organization: "ballerina", Package: "http"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePackageSelector(test.input)
			if err != nil {
				t.Fatalf("ParsePackageSelector() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("ParsePackageSelector() = %#v, want %#v", got, test.want)
			}
			if got.String() != test.want.String() {
				t.Fatalf("PackageSelector.String() = %q, want %q", got.String(), test.want.String())
			}
		})
	}
}

func TestParsePackageSelectorRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", " ", "/http", "ballerina/", "ballerina/http/client", "ballerina /http", "ballerina/h ttp"} {
		t.Run(input, func(t *testing.T) {
			if _, err := ParsePackageSelector(input); err == nil {
				t.Fatalf("ParsePackageSelector(%q) succeeded, want error", input)
			}
		})
	}
}

func TestUnavailableDocumentationService(t *testing.T) {
	service := NewUnavailableDocumentationService()
	if _, err := service.Documentation(t.Context(), PackageSelector{Package: "http"}); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("Documentation() error = %v, want ErrBackendUnavailable", err)
	}
}
