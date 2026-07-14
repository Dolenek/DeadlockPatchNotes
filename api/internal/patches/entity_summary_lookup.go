package patches

func findHeroSummary(details []PatchDetail, targetSlug string) (HeroSummary, bool) {
	for _, summary := range buildHeroList(details).Items {
		if summary.Slug == targetSlug {
			return summary, true
		}
	}
	return HeroSummary{}, false
}

func findItemSummary(details []PatchDetail, targetSlug string) (ItemSummary, bool) {
	for _, summary := range buildItemList(details).Items {
		if summary.Slug == targetSlug {
			return summary, true
		}
	}
	return ItemSummary{}, false
}

func findSpellSummary(details []PatchDetail, targetSlug string) (SpellSummary, bool) {
	for _, summary := range buildSpellList(details).Items {
		if summary.Slug == targetSlug {
			return summary, true
		}
	}
	return SpellSummary{}, false
}
