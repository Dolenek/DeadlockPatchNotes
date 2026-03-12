package patches

// Repository abstracts patch storage for HTTP handlers.
type Repository interface {
	List(page, limit int) (PatchListResponse, error)
	GetBySlug(slug string) (PatchDetail, error)
	ListHeroes() (HeroListResponse, error)
	GetHeroChanges(query HeroChangesQuery) (HeroChangesResponse, error)
	ListItems() (ItemListResponse, error)
	GetItemChanges(query ItemChangesQuery) (ItemChangesResponse, error)
	ListSpells() (SpellListResponse, error)
	GetSpellChanges(query SpellChangesQuery) (SpellChangesResponse, error)
}
