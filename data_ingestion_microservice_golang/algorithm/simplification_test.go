package algorithm

import (
	"testing"

	"data-ingestion-microservice/types"
)

func TestNewRouteSimplifier(t *testing.T) {
	tolerance := 0.001
	simplifier := NewRouteSimplifier(tolerance)
	
	if simplifier.GetTolerance() != tolerance {
		t.Errorf("Expected tolerance %f, got %f", tolerance, simplifier.GetTolerance())
	}
}

func TestSimplifyRoute_EmptyRoute(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	locations := []types.Location{}
	
	result, err := simplifier.SimplifyRoute(locations)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d points", len(result))
	}
}

func TestSimplifyRoute_TwoPoints(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	locations := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 1.0, Longitude: 1.0},
	}
	
	result, err := simplifier.SimplifyRoute(locations)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if len(result) != 2 {
		t.Errorf("Expected 2 points, got %d", len(result))
	}
	
	if result[0] != locations[0] || result[1] != locations[1] {
		t.Errorf("Expected same points as input for 2-point route")
	}
}

func TestSimplifyRoute_StraightLine(t *testing.T) {
	simplifier := NewRouteSimplifier(0.1)
	
	// Create a straight line with 5 points
	locations := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 1.0, Longitude: 1.0},
		{Latitude: 2.0, Longitude: 2.0},
		{Latitude: 3.0, Longitude: 3.0},
		{Latitude: 4.0, Longitude: 4.0},
	}
	
	result, err := simplifier.SimplifyRoute(locations)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// A straight line should be simplified to just start and end points
	if len(result) != 2 {
		t.Errorf("Expected 2 points for straight line, got %d", len(result))
	}
	
	if result[0] != locations[0] || result[1] != locations[len(locations)-1] {
		t.Errorf("Expected first and last points to be preserved")
	}
}

func TestSimplifyRoute_ZigZagLine(t *testing.T) {
	simplifier := NewRouteSimplifier(0.01)
	
	// Create a zig-zag line that should preserve more points
	locations := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 1.0, Longitude: 0.5},
		{Latitude: 2.0, Longitude: 0.0},
		{Latitude: 3.0, Longitude: 0.5},
		{Latitude: 4.0, Longitude: 0.0},
	}
	
	result, err := simplifier.SimplifyRoute(locations)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should preserve more points due to the zig-zag pattern
	if len(result) < 2 {
		t.Errorf("Expected at least 2 points, got %d", len(result))
	}
	
	// First and last points should always be preserved
	if result[0] != locations[0] || result[len(result)-1] != locations[len(locations)-1] {
		t.Errorf("Expected first and last points to be preserved")
	}
}

func TestGetCompressionStats(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	
	original := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 1.0, Longitude: 1.0},
		{Latitude: 2.0, Longitude: 2.0},
		{Latitude: 3.0, Longitude: 3.0},
		{Latitude: 4.0, Longitude: 4.0},
	}
	
	simplified := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 4.0, Longitude: 4.0},
	}
	
	stats := simplifier.GetCompressionStats(original, simplified)
	
	if stats.OriginalPoints != 5 {
		t.Errorf("Expected 5 original points, got %d", stats.OriginalPoints)
	}
	
	if stats.SimplifiedPoints != 2 {
		t.Errorf("Expected 2 simplified points, got %d", stats.SimplifiedPoints)
	}
	
	expectedRatio := 2.0 / 5.0
	if stats.CompressionRatio != expectedRatio {
		t.Errorf("Expected compression ratio %f, got %f", expectedRatio, stats.CompressionRatio)
	}
	
	if stats.PointsRemoved != 3 {
		t.Errorf("Expected 3 points removed, got %d", stats.PointsRemoved)
	}
	
	expectedReduction := 60.0 // (1 - 0.4) * 100
	if stats.ReductionPercent != expectedReduction {
		t.Errorf("Expected reduction percent %f, got %f", expectedReduction, stats.ReductionPercent)
	}
}

func TestSetTolerance(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	
	newTolerance := 0.005
	simplifier.SetTolerance(newTolerance)
	
	if simplifier.GetTolerance() != newTolerance {
		t.Errorf("Expected tolerance %f, got %f", newTolerance, simplifier.GetTolerance())
	}
}

func TestPerpendicularDistance(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	
	// Test perpendicular distance calculation
	point := Point{X: 1.0, Y: 1.0}
	lineStart := Point{X: 0.0, Y: 0.0}
	lineEnd := Point{X: 2.0, Y: 0.0}
	
	distance := simplifier.perpendicularDistance(point, lineStart, lineEnd)
	
	// The perpendicular distance from (1,1) to line from (0,0) to (2,0) should be 1.0
	if distance != 1.0 {
		t.Errorf("Expected perpendicular distance 1.0, got %f", distance)
	}
}

func TestDistance(t *testing.T) {
	simplifier := NewRouteSimplifier(0.001)
	
	p1 := Point{X: 0.0, Y: 0.0}
	p2 := Point{X: 3.0, Y: 4.0}
	
	distance := simplifier.distance(p1, p2)
	
	// Distance between (0,0) and (3,4) should be 5.0 (3-4-5 triangle)
	if distance != 5.0 {
		t.Errorf("Expected distance 5.0, got %f", distance)
	}
}

// Benchmark tests
func BenchmarkSimplifyRoute_100Points(b *testing.B) {
	simplifier := NewRouteSimplifier(0.001)
	
	// Generate 100 points in a straight line
	locations := make([]types.Location, 100)
	for i := 0; i < 100; i++ {
		locations[i] = types.Location{
			Latitude:  float64(i),
			Longitude: float64(i),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := simplifier.SimplifyRoute(locations)
		if err != nil {
			b.Fatalf("Error in simplification: %v", err)
		}
	}
}

func BenchmarkSimplifyRoute_1000Points(b *testing.B) {
	simplifier := NewRouteSimplifier(0.001)
	
	// Generate 1000 points in a straight line
	locations := make([]types.Location, 1000)
	for i := 0; i < 1000; i++ {
		locations[i] = types.Location{
			Latitude:  float64(i),
			Longitude: float64(i),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := simplifier.SimplifyRoute(locations)
		if err != nil {
			b.Fatalf("Error in simplification: %v", err)
		}
	}
} 