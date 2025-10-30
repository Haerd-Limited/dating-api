package utils

import "math"

// CalculateDistanceKm calculates the distance between two points on Earth using the Haversine formula
// Returns the distance in kilometers
func CalculateDistanceKm(lat1, lon1, lat2, lon2 float64) int {
	const earthRadiusKm = 6371.0 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadiusKm * c

	// Round to nearest kilometer
	return int(math.Round(distance))
}
