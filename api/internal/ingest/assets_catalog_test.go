package ingest

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestLoadAssetCatalogBuildsAliasesAndAbilityOwnership(t *testing.T) {
	heroes := `[
		{"id":1,"name":"The Doorman","images":{"icon_image_small":"door.png"}},
		{"id":2,"name":"Abrams","images":{"icon_image_small":"abrams.png"}}
	]`
	items := `[
		{"name":"Stalker","type":"weapon","shop_image":"stalker.png"},
		{"name":"Shoulder Charge","type":"ability","hero":1,"image":"charge.png"},
		{"name":"Shared Name","type":"ability","hero":1,"image":"one.png"},
		{"name":"Shared Name","type":"ability","hero":2,"image":"two.png"},
		{"name":"Unknown Ability","type":"ability","hero":999,"image":"unknown.png"}
	]`
	client := assetCatalogClient(map[string]testHTTPResponse{
		assetsHeroesURL: {status: http.StatusOK, body: heroes},
		assetsItemsURL:  {status: http.StatusOK, body: items},
	})

	catalog, err := LoadAssetCatalog(t.Context(), client)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	assertCatalogResolutions(t, catalog)
}

func TestLoadAssetCatalogReportsHTTPAndJSONFailures(t *testing.T) {
	testCases := []struct {
		name      string
		responses map[string]testHTTPResponse
		errorPart string
	}{
		{name: "heroes HTTP", responses: map[string]testHTTPResponse{assetsHeroesURL: {status: 503}}, errorPart: "status 503"},
		{name: "heroes JSON", responses: map[string]testHTTPResponse{assetsHeroesURL: {status: 200, body: "{"}}, errorPart: "decode"},
		{name: "items JSON", responses: map[string]testHTTPResponse{
			assetsHeroesURL: {status: 200, body: "[]"}, assetsItemsURL: {status: 200, body: "{"},
		}, errorPart: assetsItemsURL},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := LoadAssetCatalog(t.Context(), assetCatalogClient(testCase.responses))
			if err == nil || !strings.Contains(err.Error(), testCase.errorPart) {
				t.Fatalf("expected error containing %q, got %v", testCase.errorPart, err)
			}
		})
	}
}

func assertCatalogResolutions(t *testing.T, catalog *AssetCatalog) {
	t.Helper()
	if hero, ok := catalog.resolveHero("The Doorman"); !ok || hero.ID != 1 {
		t.Fatalf("hero alias did not resolve: %+v, %v", hero, ok)
	}
	if item, ok := catalog.resolveItem("Backstabber", ""); !ok || item.Name != "Stalker" {
		t.Fatalf("item alias did not resolve: %+v, %v", item, ok)
	}
	if owner, ok := catalog.resolveUniqueAbility("Shoulder Charge"); !ok || owner.HeroName != "Doorman" {
		t.Fatalf("unique ability owner did not resolve: %+v, %v", owner, ok)
	}
	if _, ok := catalog.resolveUniqueAbility("Shared Name"); ok {
		t.Fatal("ambiguous ability unexpectedly resolved")
	}
	if abilities := catalog.heroAbilities("Doorman"); len(abilities) < 2 {
		t.Fatalf("expected indexed Doorman abilities, got %+v", abilities)
	}
}

type testHTTPResponse struct {
	status int
	body   string
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func assetCatalogClient(responses map[string]testHTTPResponse) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		configured, ok := responses[request.URL.String()]
		if !ok {
			configured = testHTTPResponse{status: http.StatusNotFound}
		}
		return &http.Response{
			StatusCode: configured.status,
			Body:       io.NopCloser(strings.NewReader(configured.body)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
}
