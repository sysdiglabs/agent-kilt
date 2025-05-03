package kilt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetParameterName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		// No changes if there are no non-alphanumeric chars
		{
			name:     `SOLONGANDTHANKSFORALLTHEFISH12345`,
			expected: `SOLONGANDTHANKSFORALLTHEFISH12345`,
		},
		{
			name:     `solongandthanksforallthefish12345`,
			expected: `solongandthanksforallthefish12345`,
		},
		{
			name:     `soLongAndThanksForAllTheFish12345`,
			expected: `soLongAndThanksForAllTheFish12345`,
		},
		// Tries to make the parameter name more readable if there are non-alphanumeric chars
		{
			name:     `SOLONGANDTHANKSFORALLTHEFISH_`,
			expected: `solongandthanksforallthefish`,
		},
		{
			name:     `SOLONG_ANDTHANKSFORALLTHEFISH`,
			expected: `solongAndthanksforallthefish`,
		},
		{
			name:     `SO_LONG_AND_THANKS_FOR_ALL_THE_FISH`,
			expected: `soLongAndThanksForAllTheFish`,
		},
		{
			name:     `_SO_LONG_AND_THANKS_FOR_ALL_THE_FISH_`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		{
			name:     `__SO__LONG__AND__THANKS__FOR__ALL__THE__FISH__`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		{
			name:     `solongandthanksforallthefish_`,
			expected: `solongandthanksforallthefish`,
		},
		{
			name:     `solong_andthanksforallthefish`,
			expected: `solongAndthanksforallthefish`,
		},
		{
			name:     `so_long_and_thanks_for_all_the_fish`,
			expected: `soLongAndThanksForAllTheFish`,
		},
		{
			name:     `_so_long_and_thanks_for_all_the_fish_`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		{
			name:     `__so__long__and__thanks__for__all__the__fish__`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		{
			name:     `soLong_AndThanksForAllTheFish`,
			expected: `solongAndthanksforallthefish`,
		},
		{
			name:     `so_Long_And_Thanks_For_All_The_Fish`,
			expected: `soLongAndThanksForAllTheFish`,
		},
		{
			name:     `_so_Long_And_Thanks_For_All_The_Fish_`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		{
			name:     `__so__Long__And__Thanks__For__All__The__Fish__`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
		// Won't happen, actually
		{
			name:     `soLong-ANDTHANKS_forAllTheFish___`,
			expected: `solongAndthanksForallthefish`,
		},
		{
			name:     `soLong-ANDTHANKS-forAllTheFish!!!`,
			expected: `solongAndthanksForallthefish`,
		},
		{
			name:     `soLong-ANDTHANKS-forAllTheFish!!!`,
			expected: `solongAndthanksForallthefish`,
		},
		{
			name:     `soLongAndThanksForAllTheFish!!!`,
			expected: `solongandthanksforallthefish`,
		},
		{
			name:     `***so___Long---And!!!Thanks???For+++All***The:::Fish|||`,
			expected: `SoLongAndThanksForAllTheFish`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getParameterName(tc.name))
		})
	}
}
