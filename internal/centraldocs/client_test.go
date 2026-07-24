package centraldocs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientScrapeModule(t *testing.T) {
	page := `{"props":{"pageProps":{"propsData":{"props":{"module":{"id":"demo","summary":"Module summary","description":"Module overview","orgName":"acme","version":"1.0.0","isDefaultModule":true,"relatedModules":[{"id":"demo","isDefaultModule":true},{"id":"demo.extra","isDefaultModule":false}],"functions":[{"name":"ping","description":"Pings.","parameters":[],"returnParameters":[]}]},"packageData":{"organization":"acme","name":"demo","version":"1.0.0","summary":"Package summary","readme":"# Demo","defaultModuleName":"demo"}}}}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acme/demo/1.0.0" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fmt.Fprintf(w, `<html><script id="__NEXT_DATA__" type="application/json">%s</script></html>`, page)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ScrapeModule(t.Context(), "acme", "demo", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if result.Module.Name != "demo" || len(result.Module.Functions) != 1 || result.Module.Functions[0].Signature != "function ping()" {
		t.Fatalf("result = %#v", result)
	}
	encoded, err := json.Marshal(result.Module)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" {
		t.Fatal("empty JSON")
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"functions", "resources", "classes", "clients", "records", "types"} {
		if object[key] == nil {
			t.Errorf("%s encoded as null", key)
		}
	}
}

func TestExtractNextDataErrors(t *testing.T) {
	for name, input := range map[string]string{
		"missing":   `<html></html>`,
		"empty":     `<script id="__NEXT_DATA__"></script>`,
		"duplicate": `<script id="__NEXT_DATA__">{}</script><script id="__NEXT_DATA__">{}</script>`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := extractNextData(strings.NewReader(input)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
