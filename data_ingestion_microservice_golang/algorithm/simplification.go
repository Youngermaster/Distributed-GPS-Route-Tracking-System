package algorithm

import (
	"math"

	"data-ingestion-microservice/types"
)

// RouteSimplifier handles route simplification using various algorithms
type RouteSimplifier struct {
	tolerance float64
}

// NewRouteSimplifier creates a new route simplifier with the given tolerance
func NewRouteSimplifier(tolerance float64) *RouteSimplifier {
	return &RouteSimplifier{
		tolerance: tolerance,
	}
}

// Point represents a 2D point for algorithm calculations
type Point struct {
	X, Y float64
}

// SimplifyRoute simplifies a route using the Douglas-Peucker algorithm
func (rs *RouteSimplifier) SimplifyRoute(locations []types.Location) ([]types.Location, error) {
	if len(locations) <= 2 {
		return locations, nil
	}

	// Convert locations to points
	points := make([]Point, len(locations))
	for i, loc := range locations {
		points[i] = Point{X: loc.Longitude, Y: loc.Latitude}
	}

	// Apply Douglas-Peucker algorithm
	simplified := rs.douglasPeucker(points, rs.tolerance)

	// Convert back to Location structs
	result := make([]types.Location, len(simplified))
	for i, point := range simplified {
		result[i] = types.Location{
			Longitude: point.X,
			Latitude:  point.Y,
		}
	}

	return result, nil
}

// douglasPeucker implements the Ramer-Douglas-Peucker algorithm
func (rs *RouteSimplifier) douglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// Find the point with the maximum distance from the line segment
	// defined by the first and last points
	maxDistance := 0.0
	maxIndex := 0
	start := points[0]
	end := points[len(points)-1]

	for i := 1; i < len(points)-1; i++ {
		distance := rs.perpendicularDistance(points[i], start, end)
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	// If the maximum distance is greater than tolerance, recursively simplify
	if maxDistance > tolerance {
		// Recursive call on the first part
		firstPart := rs.douglasPeucker(points[:maxIndex+1], tolerance)
		// Recursive call on the second part
		secondPart := rs.douglasPeucker(points[maxIndex:], tolerance)

		// Combine results (remove duplicate point at the junction)
		result := make([]Point, len(firstPart)+len(secondPart)-1)
		copy(result, firstPart)
		copy(result[len(firstPart):], secondPart[1:])
		return result
	}

	// If no point is farther than tolerance, return only start and end points
	return []Point{start, end}
}

// perpendicularDistance calculates the perpendicular distance from a point to a line segment
func (rs *RouteSimplifier) perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	// Calculate the area of the triangle formed by the three points
	// using the cross product, then divide by the length of the base
	area := math.Abs(lineStart.X*(lineEnd.Y-point.Y) + lineEnd.X*(point.Y-lineStart.Y) + point.X*(lineStart.Y-lineEnd.Y))

	// Calculate the length of the base (line segment)
	base := rs.distance(lineStart, lineEnd)

	if base == 0 {
		return rs.distance(point, lineStart)
	}

	return area / base
}

// distance calculates the Euclidean distance between two points
func (rs *RouteSimplifier) distance(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// GetCompressionStats returns statistics about the compression
func (rs *RouteSimplifier) GetCompressionStats(original, simplified []types.Location) CompressionStats {
	compressionRatio := float64(len(simplified)) / float64(len(original))
	
	return CompressionStats{
		OriginalPoints:    len(original),
		SimplifiedPoints:  len(simplified),
		CompressionRatio:  compressionRatio,
		PointsRemoved:     len(original) - len(simplified),
		ReductionPercent:  (1 - compressionRatio) * 100,
	}
}

// CompressionStats holds statistics about route compression
type CompressionStats struct {
	OriginalPoints    int     `json:"originalPoints"`
	SimplifiedPoints  int     `json:"simplifiedPoints"`
	CompressionRatio  float64 `json:"compressionRatio"`
	PointsRemoved     int     `json:"pointsRemoved"`
	ReductionPercent  float64 `json:"reductionPercent"`
}

// SetTolerance updates the tolerance value
func (rs *RouteSimplifier) SetTolerance(tolerance float64) {
	rs.tolerance = tolerance
}

// GetTolerance returns the current tolerance value
func (rs *RouteSimplifier) GetTolerance() float64 {
	return rs.tolerance
} 