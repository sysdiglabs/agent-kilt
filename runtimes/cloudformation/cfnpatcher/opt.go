package cfnpatcher

import (
	"strings"
)

const kiltIgnoreTag = "kilt-ignore"
const kiltIncludeTag = "kilt-include"
const kiltIgnoreContainersTag = "kilt-ignore-containers"
const kiltIncludeContainersTag = "kilt-include-containers"

var optTagKeys = []string{kiltIgnoreTag, kiltIncludeTag, kiltIgnoreContainersTag, kiltIncludeContainersTag}

func isOptTagKey(key string) bool {
	for _, v := range optTagKeys {
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
	_, included := tags[kiltIncludeTag]
	return &InstrumentationHints{
		IgnoreContainersNamed:  extractContainersFromTag(tags, kiltIgnoreContainersTag),
		IncludeContainersNamed: extractContainersFromTag(tags, kiltIncludeContainersTag),
		HasGlobalInclude:       included,
	}
}

func isIgnored(tags map[string]string, isOptIn bool) bool {
	_, included := tags[kiltIncludeTag]
	_, ignored := tags[kiltIgnoreTag]
	_, hasNamedContainerIncluded := tags[kiltIncludeContainersTag]

	return !((isOptIn && (included || hasNamedContainerIncluded)) || (!isOptIn && !ignored))
}
