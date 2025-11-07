package domain

// DiscoverPreferenceUpdate represents preference values to persist for a user.
type DiscoverPreferenceUpdate struct {
	DistanceKM         *int
	MinAge             *int
	MaxAge             *int
	DatingIntentionIDs []int16
	ReligionIDs        []int16
	EthnicityIDs       []int16
}

// StoredDiscoverPreferences encapsulates persisted preference values.
type StoredDiscoverPreferences struct {
	DistanceKM         *int
	MinAge             *int
	MaxAge             *int
	DatingIntentionIDs []int16
	ReligionIDs        []int16
	EthnicityIDs       []int16
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
		len(p.EthnicityIDs) > 0
}

// NewPreferenceUpdateFromFilters creates an update payload from discover filters.
func NewPreferenceUpdateFromFilters(filters *DiscoverFilters) *DiscoverPreferenceUpdate {
	if filters == nil || filters.IsEmpty() {
		return nil
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
		len(u.EthnicityIDs) > 0
}
