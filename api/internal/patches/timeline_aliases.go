package patches

import "strings"

var timelineHeroAlias = map[string]string{
	"doorman": "the doorman",
	"vindcita": "vindicta",
}

var timelineAbilityAlias = map[string]map[string][]string{
	"bebop": {
		"grapple arm": []string{"hook"},
		"hyper beam":  []string{"hyperbeam"},
		"exploding uppercut": []string{"uppercut"},
	},
}

func canonicalTimelineHeroName(name string) string {
	key := canonicalTimelineHeroKey(name)
	if key == "doorman" {
		return "Doorman"
	}
	return strings.TrimSpace(name)
}

func canonicalTimelineHeroKey(name string) string {
	key := normalizeLookupKey(name)
	if alias, ok := timelineHeroAlias[key]; ok {
		key = normalizeLookupKey(alias)
	}
	if key == "the doorman" {
		return "doorman"
	}
	return key
}
