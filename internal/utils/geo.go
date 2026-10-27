package utils

import (
	"math"
)

// HaversineDistance calculates the great-circle distance between two GPS coordinates in kilometers
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// If coordinates are identical or invalid, return 0
	if lat1 == 0 && lon1 == 0 {
		return 0
	}
	if lat2 == 0 && lon2 == 0 {
		return 0
	}
	if lat1 == lat2 && lon1 == lon2 {
		return 0
	}

	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return math.Round(earthRadiusKm*c*100) / 100 // rounded to 2 decimal places
}
