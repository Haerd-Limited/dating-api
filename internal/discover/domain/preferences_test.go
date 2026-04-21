package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPreferenceUpdateFromFilters(t *testing.T) {
	minAge := 25
	maxAge := 35
	distance := 50

	filters := &DiscoverFilters{
		AgeRange: &AgeRangeFilter{
			MinAge: &minAge,
			MaxAge: &maxAge,
		},
		Distance: &DistanceFilter{
			MaxDistanceKM: &distance,
		},
		Religions: &ReligionsFilter{
			ReligionIDs: []int16{3},
		},
		Ethnicities: &EthnicitiesFilter{
			EthnicityIDs: []int16{4, 5},
		},
	}

	update := NewPreferenceUpdateFromFilters(filters)
	require.NotNil(t, update)

	require.NotNil(t, update.DistanceKM)
	assert.Equal(t, distance, *update.DistanceKM)

	require.NotNil(t, update.MinAge)
	assert.Equal(t, minAge, *update.MinAge)

	require.NotNil(t, update.MaxAge)
	assert.Equal(t, maxAge, *update.MaxAge)

	assert.Equal(t, []int16{3}, update.ReligionIDs)
	assert.Equal(t, []int16{4, 5}, update.EthnicityIDs)
}

func TestNewPreferenceUpdateFromFiltersReturnsNilWhenEmpty(t *testing.T) {
	assert.Nil(t, NewPreferenceUpdateFromFilters(nil))
}

func TestNewPreferenceUpdateFromFiltersReturnsClearAllWhenFiltersEmpty(t *testing.T) {
	// Explicitly empty filters (e.g. user cleared all) should persist as clear-all so restart doesn't revert
	update := NewPreferenceUpdateFromFilters(&DiscoverFilters{})
	require.NotNil(t, update)
	assert.True(t, update.ClearAll)
}

func TestStoredDiscoverPreferencesHasAnyPreference(t *testing.T) {
	assert.False(t, (*StoredDiscoverPreferences)(nil).HasAnyPreference())
	assert.False(t, (&StoredDiscoverPreferences{}).HasAnyPreference())

	distance := 20
	prefs := &StoredDiscoverPreferences{
		DistanceKM: &distance,
	}

	assert.True(t, prefs.HasAnyPreference())
}
