package patches

import "context"

// Repository abstracts patch storage for HTTP handlers.
type Repository interface {
	List(ctx context.Context, page, limit int) (PatchListResponse, error)
	GetBySlug(ctx context.Context, slug string) (PatchDetail, error)
	ListHeroes(ctx context.Context) (HeroListResponse, error)
	GetHeroChanges(ctx context.Context, query HeroChangesQuery) (HeroChangesResponse, error)
	ListItems(ctx context.Context) (ItemListResponse, error)
	GetItemChanges(ctx context.Context, query ItemChangesQuery) (ItemChangesResponse, error)
	ListSpells(ctx context.Context) (SpellListResponse, error)
	GetSpellChanges(ctx context.Context, query SpellChangesQuery) (SpellChangesResponse, error)
}
