package patches

// Repository abstracts patch storage for HTTP handlers.
type Repository interface {
	List(page, limit int) ListResponse
	GetBySlug(slug string) (PatchDetail, error)
	ListHeroes() HeroListResponse
	GetHeroChanges(query HeroChangesQuery) (HeroChangesResponse, error)
	ListItems() ItemListResponse
	GetItemChanges(query ItemChangesQuery) (ItemChangesResponse, error)
	ListSpells() SpellListResponse
	GetSpellChanges(query SpellChangesQuery) (SpellChangesResponse, error)
}
