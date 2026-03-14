package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"deadlockpatchnotes/api/internal/structuredparse"
)

const (
	assetsHeroesURL = "https://assets.deadlock-api.com/v2/heroes"
	assetsItemsURL  = "https://assets.deadlock-api.com/v2/items"
)

var (
	itemAlias = map[string]string{
		"backstabber": "stalker",
	}
	itemRenameRegex = regexp.MustCompile(`(?i)^renamed to\s+(.+?)[.]*$`)
)

type heroImages struct {
	IconImageSmall string `json:"icon_image_small"`
}

type heroAsset struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Images heroImages `json:"images"`
}

type itemAsset struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Image         string `json:"image"`
	ImageWebP     string `json:"image_webp"`
	ShopImage     string `json:"shop_image"`
	ShopImageWebP string `json:"shop_image_webp"`
	HeroID        int    `json:"hero"`
}

type abilityRef struct {
	Name      string
	NormName  string
	Image     string
	ImageWebP string
}

type AssetCatalog struct {
	heroesByNorm     map[string]heroAsset
	heroByID         map[int]heroAsset
	itemsByNorm      map[string]itemAsset
	abilitiesByHero  map[string][]abilityRef
}

func LoadAssetCatalog(ctx context.Context, client *http.Client) (*AssetCatalog, error) {
	var heroes []heroAsset
	if err := fetchJSON(ctx, client, assetsHeroesURL, &heroes); err != nil {
		return nil, err
	}

	var items []itemAsset
	if err := fetchJSON(ctx, client, assetsItemsURL, &items); err != nil {
		return nil, err
	}

	catalog := &AssetCatalog{
		heroesByNorm:    make(map[string]heroAsset, len(heroes)*2),
		heroByID:        make(map[int]heroAsset, len(heroes)),
		itemsByNorm:     make(map[string]itemAsset, len(items)),
		abilitiesByHero: map[string][]abilityRef{},
	}

	for _, hero := range heroes {
		key := structuredparse.NormalizeLookupKey(hero.Name)
		catalog.heroesByNorm[key] = hero
		catalog.heroesByNorm[strings.TrimPrefix(key, "the ")] = hero
		catalog.heroByID[hero.ID] = hero
	}

	for _, item := range items {
		catalog.itemsByNorm[structuredparse.NormalizeLookupKey(item.Name)] = item
		if item.Type != "ability" || item.HeroID == 0 {
			continue
		}
		hero, ok := catalog.heroByID[item.HeroID]
		if !ok {
			continue
		}
		heroKey := structuredparse.CanonicalHeroKey(hero.Name)
		abilities := structuredparse.ExpandAbilityAliases(heroKey, []structuredparse.AbilityRef{
			{
				Name:            item.Name,
				IconFallbackURL: firstNonEmpty(item.Image, item.ImageWebP),
			},
		})
		for _, ability := range abilities {
			catalog.abilitiesByHero[heroKey] = append(catalog.abilitiesByHero[heroKey], abilityRef{
				Name:      ability.Name,
				NormName:  ability.NormName,
				Image:     ability.IconFallbackURL,
				ImageWebP: ability.IconFallbackURL,
			})
		}
	}

	return catalog, nil
}

func fetchJSON(ctx context.Context, client *http.Client, sourceURL string, out any) error {
	raw, err := fetchText(ctx, client, sourceURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return fmt.Errorf("decode %s: %w", sourceURL, err)
	}
	return nil
}

func (c *AssetCatalog) resolveHero(name string) (heroAsset, bool) {
	if c == nil {
		return heroAsset{}, false
	}
	key := structuredparse.ResolveHeroAlias(name)
	hero, ok := c.heroesByNorm[key]
	return hero, ok
}

func (c *AssetCatalog) resolveItem(name, changeText string) (itemAsset, bool) {
	if c == nil {
		return itemAsset{}, false
	}
	key := structuredparse.NormalizeLookupKey(name)
	if item, ok := c.itemsByNorm[key]; ok {
		return item, true
	}
	if alias, ok := itemAlias[key]; ok {
		if item, ok := c.itemsByNorm[structuredparse.NormalizeLookupKey(alias)]; ok {
			return item, true
		}
	}
	if match := itemRenameRegex.FindStringSubmatch(strings.TrimSpace(changeText)); len(match) == 2 {
		if item, ok := c.itemsByNorm[structuredparse.NormalizeLookupKey(match[1])]; ok {
			return item, true
		}
	}
	return itemAsset{}, false
}

func (c *AssetCatalog) heroAbilities(heroName string) []abilityRef {
	if c == nil {
		return nil
	}

	key := structuredparse.CanonicalHeroKey(heroName)
	if abilities, ok := c.abilitiesByHero[key]; ok {
		return abilities
	}
	return nil
}

func resolveHeroDisplayName(rawPrefix string, hero heroAsset) string {
	if displayName := structuredparse.CanonicalHeroDisplayName(rawPrefix); displayName != "" {
		if structuredparse.CanonicalHeroKey(displayName) == structuredparse.CanonicalHeroKey(hero.Name) {
			return displayName
		}
	}
	return structuredparse.CanonicalHeroDisplayName(hero.Name)
}

func itemImage(item itemAsset) string {
	for _, candidate := range []string{item.ShopImage, item.ShopImageWebP, item.Image, item.ImageWebP} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}
