package cfnpatcher

import (
	"strings"
)

const KiltIgnoreTag = "kilt-ignore"
const KiltIncludeTag = "kilt-include"
const KiltIgnoreContainersTag = "kilt-ignore-containers"
const KiltIncludeContainersTag = "kilt-include-containers"

var OptTagKeys = []string{KiltIgnoreTag, KiltIncludeTag, KiltIgnoreContainersTag, KiltIncludeContainersTag}

func isOptTagKey(key string) bool {
	for _, v := range OptTagKeys {
		if key == v {
			return true
		}
	}
	return false
}

func extractContainersFromTag(tags map[string]string, tag string) []string {
	containers := make([]string, 0)
	containerList, hasIgnores := tags[tag]
	if hasIgnores {
		containers = strings.Split(containerList, ":")
	}
	return containers
}

func extractHintsFromTags(tags map[string]string) *InstrumentationHints {
	_, included := tags[KiltIncludeTag]
	return &InstrumentationHints{
		IgnoreContainersNamed:  extractContainersFromTag(tags, KiltIgnoreContainersTag),
		IncludeContainersNamed: extractContainersFromTag(tags, KiltIncludeContainersTag),
		HasGlobalInclude:       included,
	}
}

func isIgnored(tags map[string]string, isOptIn bool) bool {
	_, included := tags[KiltIncludeTag]
	_, ignored := tags[KiltIgnoreTag]
	_, hasNamedContainerIncluded := tags[KiltIncludeContainersTag]

	return !((isOptIn && (included || hasNamedContainerIncluded)) || (!isOptIn && !ignored))
}
