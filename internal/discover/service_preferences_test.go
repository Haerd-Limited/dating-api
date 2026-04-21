package discover

import (
	"testing"

	"github.com/stretchr/testify/assert"

	discoverdomain "github.com/Haerd-Limited/dating-api/internal/discover/domain"
)

func TestNewPreferenceMatcherReturnsNilWhenNoPreferences(t *testing.T) {
	assert.Nil(t, newPreferenceMatcher(nil))
	assert.Nil(t, newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{}))
}

func TestPreferenceMatcherMatchesAnyPreference(t *testing.T) {
	minAge := 25
	maxAge := 35
	distance := 30

	matcher := newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{
		MinAge:       &minAge,
		MaxAge:       &maxAge,
		DistanceKM:   &distance,
		ReligionIDs:  []int16{3},
		EthnicityIDs: []int16{4},
	})

	if matcher == nil {
		t.Fatal("expected matcher to be created")
	}

	assert.True(t, matcher.matchesAny(28, 10, nil, nil, nil), "distance preference should match")
	assert.True(t, matcher.matchesAny(30, 100, nil, nil, nil), "age preference should match")

	religionID := int16(3)
	assert.True(t, matcher.matchesAny(50, 100, &religionID, nil, nil), "religion preference should match")

	assert.True(t, matcher.matchesAny(50, 100, nil, nil, []int16{4}), "ethnicity preference should match")

	assert.False(t, matcher.matchesAny(50, 100, nil, nil, []int16{5}), "should not match when no preferences satisfied")
}

func TestPreferenceMatcherMatchesAllPreferences(t *testing.T) {
	minAge := 25
	maxAge := 35
	distance := 30

	matcher := newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{
		MinAge:       &minAge,
		MaxAge:       &maxAge,
		DistanceKM:   &distance,
		ReligionIDs:  []int16{3},
		EthnicityIDs: []int16{4},
	})

	if matcher == nil {
		t.Fatal("expected matcher to be created")
	}

	religionID := int16(3)

	assert.True(t, matcher.matchesAll(30, 10, &religionID, nil, []int16{4}), "should match when all preferences satisfied")
	assert.False(t, matcher.matchesAll(30, 100, &religionID, nil, []int16{4}), "should fail distance preference")
	assert.False(t, matcher.matchesAll(36, 10, &religionID, nil, []int16{4}), "should fail age preference")
	assert.False(t, matcher.matchesAll(30, 10, nil, nil, []int16{4}), "should fail religion preference")
	assert.False(t, matcher.matchesAll(30, 10, &religionID, nil, []int16{5}), "should fail ethnicity preference")
}

func TestPreferenceMatcherRequiresEthnicity(t *testing.T) {
	matcher := newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{
		EthnicityIDs: []int16{5},
	})

	assert.True(t, matcher.requiresEthnicity())

	noEthnicity := newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{
		DistanceKM: ptrToInt(10),
	})

	assert.False(t, noEthnicity.requiresEthnicity())
}

func ptrToInt(value int) *int {
	return &value
}
