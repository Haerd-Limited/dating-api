package domain

// FilterOperator represents how filters should be combined
type FilterOperator string

const (
	FilterOperatorAND FilterOperator = "AND"
	FilterOperatorOR  FilterOperator = "OR"
)

// SeekGenderFilter values for discover feed: show only Male, only Female, or Both.
const (
	SeekGenderMale   = "Male"
	SeekGenderFemale = "Female"
	SeekGenderBoth   = "Both"
)

// DiscoverFilters represents all available filters for the discover feed
type DiscoverFilters struct {
	AgeRange    *AgeRangeFilter    `json:"age_range,omitempty"`
	Distance    *DistanceFilter    `json:"distance,omitempty"`
	Religions   *ReligionsFilter   `json:"religions,omitempty"`
	Sexualities *SexualitiesFilter `json:"sexualities,omitempty"`
	Ethnicities *EthnicitiesFilter `json:"ethnicities,omitempty"`
	SeekGender  *string            `json:"seek_gender,omitempty"` // "Male", "Female", or "Both"
	Operator    FilterOperator     `json:"operator"`              // How to combine filters (AND/OR)
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

// ReligionsFilter filters profiles by religions
type ReligionsFilter struct {
	ReligionIDs []int16 `json:"religion_ids,omitempty"`
}

// SexualitiesFilter filters profiles by sexualities
type SexualitiesFilter struct {
	SexualityIDs []int16 `json:"sexuality_ids,omitempty"`
}

// EthnicitiesFilter filters profiles by ethnicities
type EthnicitiesFilter struct {
	EthnicityIDs []int16 `json:"ethnicity_ids,omitempty"`
}

// IsEmpty returns true if no filters are set
func (f *DiscoverFilters) IsEmpty() bool {
	return f.AgeRange == nil &&
		f.Distance == nil &&
		f.Religions == nil &&
		f.Sexualities == nil &&
		f.Ethnicities == nil &&
		!f.HasSeekGenderFilter()
}

// HasSeekGenderFilter returns true if seek_gender filter is set (Male, Female, or Both).
func (f *DiscoverFilters) HasSeekGenderFilter() bool {
	if f == nil || f.SeekGender == nil {
		return false
	}

	s := *f.SeekGender

	return s == SeekGenderMale || s == SeekGenderFemale || s == SeekGenderBoth
}

// HasAgeFilter returns true if age range filter is set
func (f *DiscoverFilters) HasAgeFilter() bool {
	return f.AgeRange != nil && (f.AgeRange.MinAge != nil || f.AgeRange.MaxAge != nil)
}

// HasDistanceFilter returns true if distance filter is set
func (f *DiscoverFilters) HasDistanceFilter() bool {
	return f.Distance != nil && f.Distance.MaxDistanceKM != nil
}

// HasReligionsFilter returns true if religions filter is set
func (f *DiscoverFilters) HasReligionsFilter() bool {
	return f.Religions != nil && len(f.Religions.ReligionIDs) > 0
}

// HasSexualitiesFilter returns true if sexualities filter is set
func (f *DiscoverFilters) HasSexualitiesFilter() bool {
	return f.Sexualities != nil && len(f.Sexualities.SexualityIDs) > 0
}

// HasEthnicitiesFilter returns true if ethnicities filter is set
func (f *DiscoverFilters) HasEthnicitiesFilter() bool {
	return f.Ethnicities != nil && len(f.Ethnicities.EthnicityIDs) > 0
}
