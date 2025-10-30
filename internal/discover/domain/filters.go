package domain

// FilterOperator represents how filters should be combined
type FilterOperator string

const (
	FilterOperatorAND FilterOperator = "AND"
	FilterOperatorOR  FilterOperator = "OR"
)

// DiscoverFilters represents all available filters for the discover feed
type DiscoverFilters struct {
	AgeRange         *AgeRangeFilter         `json:"age_range,omitempty"`
	Distance         *DistanceFilter         `json:"distance,omitempty"`
	DatingIntentions *DatingIntentionsFilter `json:"dating_intentions,omitempty"`
	Religions        *ReligionsFilter        `json:"religions,omitempty"`
	Ethnicities      *EthnicitiesFilter      `json:"ethnicities,omitempty"`
	Operator         FilterOperator          `json:"operator"` // How to combine filters (AND/OR)
}

// AgeRangeFilter filters profiles by age range
type AgeRangeFilter struct {
	MinAge *int `json:"min_age,omitempty"`
	MaxAge *int `json:"max_age,omitempty"`
}

// DistanceFilter filters profiles by distance from user's location
type DistanceFilter struct {
	MaxDistanceKM *int `json:"max_distance_km,omitempty"`
}

// DatingIntentionsFilter filters profiles by dating intentions
type DatingIntentionsFilter struct {
	IntentionIDs []int16 `json:"intention_ids,omitempty"`
}

// ReligionsFilter filters profiles by religions
type ReligionsFilter struct {
	ReligionIDs []int16 `json:"religion_ids,omitempty"`
}

// EthnicitiesFilter filters profiles by ethnicities
type EthnicitiesFilter struct {
	EthnicityIDs []int16 `json:"ethnicity_ids,omitempty"`
}

// IsEmpty returns true if no filters are set
func (f *DiscoverFilters) IsEmpty() bool {
	return f.AgeRange == nil &&
		f.Distance == nil &&
		f.DatingIntentions == nil &&
		f.Religions == nil &&
		f.Ethnicities == nil
}

// HasAgeFilter returns true if age range filter is set
func (f *DiscoverFilters) HasAgeFilter() bool {
	return f.AgeRange != nil && (f.AgeRange.MinAge != nil || f.AgeRange.MaxAge != nil)
}

// HasDistanceFilter returns true if distance filter is set
func (f *DiscoverFilters) HasDistanceFilter() bool {
	return f.Distance != nil && f.Distance.MaxDistanceKM != nil
}

// HasDatingIntentionsFilter returns true if dating intentions filter is set
func (f *DiscoverFilters) HasDatingIntentionsFilter() bool {
	return f.DatingIntentions != nil && len(f.DatingIntentions.IntentionIDs) > 0
}

// HasReligionsFilter returns true if religions filter is set
func (f *DiscoverFilters) HasReligionsFilter() bool {
	return f.Religions != nil && len(f.Religions.ReligionIDs) > 0
}

// HasEthnicitiesFilter returns true if ethnicities filter is set
func (f *DiscoverFilters) HasEthnicitiesFilter() bool {
	return f.Ethnicities != nil && len(f.Ethnicities.EthnicityIDs) > 0
}
