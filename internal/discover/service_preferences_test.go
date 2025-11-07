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

func TestPreferenceMatcherMatchesIndividualPreferences(t *testing.T) {
	minAge := 25
	maxAge := 35
	distance := 30

	matcher := newPreferenceMatcher(&discoverdomain.StoredDiscoverPreferences{
		MinAge:             &minAge,
		MaxAge:             &maxAge,
		DistanceKM:         &distance,
		DatingIntentionIDs: []int16{2},
		ReligionIDs:        []int16{3},
		EthnicityIDs:       []int16{4},
	})

	if matcher == nil {
		t.Fatal("expected matcher to be created")
	}

	assert.True(t, matcher.matches(28, 10, nil, nil, nil), "distance preference should match")
	assert.True(t, matcher.matches(30, 100, nil, nil, nil), "age preference should match")

	datingID := int16(2)
	assert.True(t, matcher.matches(50, 100, &datingID, nil, nil), "dating intention preference should match")

	religionID := int16(3)
	assert.True(t, matcher.matches(50, 100, nil, &religionID, nil), "religion preference should match")

	assert.True(t, matcher.matches(50, 100, nil, nil, []int16{4}), "ethnicity preference should match")

	assert.False(t, matcher.matches(50, 100, nil, nil, []int16{5}), "should not match when no preferences satisfied")
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
