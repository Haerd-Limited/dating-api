package domain

// DiscoverPreferenceUpdate represents preference values to persist for a user.
type DiscoverPreferenceUpdate struct {
	// ClearAll when true instructs the repository to clear all discover preferences (set to null/empty).
	// Used when the user explicitly clears all filters so the cleared state is persisted.
	ClearAll           bool
	DistanceKM         *int
	MinAge             *int
	MaxAge             *int
	DatingIntentionIDs []int16
	ReligionIDs        []int16
	SexualityIDs       []int16
	EthnicityIDs       []int16
	SeekGenderIDs      []int16 // one or two gender IDs (Male, Female, or both)
}

// StoredDiscoverPreferences encapsulates persisted preference values.
type StoredDiscoverPreferences struct {
	DistanceKM         *int
	MinAge             *int
	MaxAge             *int
	DatingIntentionIDs []int16
	ReligionIDs        []int16
	SexualityIDs       []int16
	EthnicityIDs       []int16
	SeekGenderIDs      []int16
	SeekGender         string // Human-readable "Male", "Female", or "Both" derived from SeekGenderIDs (set by service when loading)
}

// HasAnyPreference returns true if the preferences contain at least one value.
func (p *StoredDiscoverPreferences) HasAnyPreference() bool {
	if p == nil {
		return false
	}

	return p.DistanceKM != nil ||
		p.MinAge != nil ||
		p.MaxAge != nil ||
		len(p.DatingIntentionIDs) > 0 ||
		len(p.ReligionIDs) > 0 ||
		len(p.SexualityIDs) > 0 ||
		len(p.EthnicityIDs) > 0 ||
		len(p.SeekGenderIDs) > 0
}

// NewPreferenceUpdateFromFilters creates an update payload from discover filters.
// When filters are explicitly empty (user cleared all), returns an update with ClearAll true so the cleared state is persisted.
func NewPreferenceUpdateFromFilters(filters *DiscoverFilters) *DiscoverPreferenceUpdate {
	if filters == nil {
		return nil
	}
	if filters.IsEmpty() {
		return &DiscoverPreferenceUpdate{ClearAll: true}
	}

	update := &DiscoverPreferenceUpdate{}

	if filters.Distance != nil && filters.Distance.MaxDistanceKM != nil {
		value := *filters.Distance.MaxDistanceKM
		update.DistanceKM = &value
	}

	if filters.AgeRange != nil {
		if filters.AgeRange.MinAge != nil {
			value := *filters.AgeRange.MinAge
			update.MinAge = &value
		}

		if filters.AgeRange.MaxAge != nil {
			value := *filters.AgeRange.MaxAge
			update.MaxAge = &value
		}
	}

	if filters.HasDatingIntentionsFilter() {
		update.DatingIntentionIDs = append([]int16{}, filters.DatingIntentions.IntentionIDs...)
	}

	if filters.HasReligionsFilter() {
		update.ReligionIDs = append([]int16{}, filters.Religions.ReligionIDs...)
	}

	if filters.HasSexualitiesFilter() {
		update.SexualityIDs = append([]int16{}, filters.Sexualities.SexualityIDs...)
	}

	if filters.HasEthnicitiesFilter() {
		update.EthnicityIDs = append([]int16{}, filters.Ethnicities.EthnicityIDs...)
	}

	if !update.hasValues() {
		return nil
	}

	return update
}

func (u *DiscoverPreferenceUpdate) hasValues() bool {
	if u == nil {
		return false
	}

	return u.DistanceKM != nil ||
		u.MinAge != nil ||
		u.MaxAge != nil ||
		len(u.DatingIntentionIDs) > 0 ||
		len(u.ReligionIDs) > 0 ||
		len(u.SexualityIDs) > 0 ||
		len(u.EthnicityIDs) > 0 ||
		len(u.SeekGenderIDs) > 0
}
