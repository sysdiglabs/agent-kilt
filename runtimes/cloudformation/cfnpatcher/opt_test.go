package cfnpatcher

import (
	"testing"
	"reflect"

    "github.com/Jeffail/gabs/v2"
	"github.com/stretchr/testify/assert"
)

func TestIsOptTagKey(t *testing.T) {
	tests := []struct {
		key string
		out bool
	}{
		{
			"kilt-ignore",
			true,
		},
		{
			"kilt-include",
			true,
		},
		{
			"kilt-ignore-containers",
			true,
		},
		{
			"kilt-include-containers",
			true,
		},
		{
			"so-long-and-thanks-for-all-the-fish",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			assert.Equal(t, test.out, isOptTagKey(test.key), "OptIn/Out key not recognized")
		})
	}
}

func TestGetOptTags(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected map[string]string
	}{
		{
			name: `no-properties`,
			json: `{
"McGuffin": {
	"Tags":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `no-tags`,
			json: `{
"Properties": {
	"Accio":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `no-opt-tags`,
			json: `{
"Properties": {
	"Tags":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		},
		{
			"Key": "TimeIsAnIllusion",
			"Value": "LunchtimeDoublySo"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `all-opt-tags`,
			json: `{
"Properties": {
	"Tags": [
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		},
		{
			"Key": "TimeIsAnIllusion",
			"Value": "LunchtimeDoublySo"
		},
		{
			"Key": "kilt-ignore",
			"Value": "nanananananaBatman"
		},
		{
			"Key": "kilt-include",
			"Value": "gimmeGimmeGimmeFriedChicken"
		},
		{
			"Key": "kilt-ignore-containers",
			"Value": "expelliarmus"
		},
		{
			"Key": "kilt-include-containers",
			"Value": "accioContainer"
		}
	]}
}`,
			expected: map[string]string{
				"kilt-ignore":             "nanananananaBatman",
				"kilt-include":            "gimmeGimmeGimmeFriedChicken",
				"kilt-ignore-containers":  "expelliarmus",
				"kilt-include-containers": "accioContainer",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonParsed, err := gabs.ParseJSON([]byte(tc.json))
			if err != nil {
				panic(err)
			}
			mm := getOptTags(jsonParsed)
			eq := reflect.DeepEqual(tc.expected, mm)
			if !eq {
				assert.Fail(t, "maps do not match")
			}
		})
	}
}
