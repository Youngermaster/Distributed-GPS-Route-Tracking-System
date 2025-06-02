package main

import (
	"math"
	"testing"
)

func TestSimplifyRoute(t *testing.T) {
	tests := []struct {
		name      string
		points    []Point
		tolerance float64
		expected  int // expected number of simplified points
	}{
		{
			name: "Empty route",
			points: []Point{},
			tolerance: 0.0001,
			expected: 0,
		},
		{
			name: "Single point",
			points: []Point{{X: 0, Y: 0}},
			tolerance: 0.0001,
			expected: 1,
		},
		{
			name: "Two points",
			points: []Point{{X: 0, Y: 0}, {X: 1, Y: 1}},
			tolerance: 0.0001,
			expected: 2,
		},
		{
			name: "Straight line - should be simplified",
			points: []Point{
				{X: 0, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 1},
			},
			tolerance: 0.1,
			expected: 2, // Only start and end points
		},
		{
			name: "Complex route with deviation",
			points: []Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0.1}, // slight deviation
				{X: 2, Y: 0.2}, // slight deviation
				{X: 3, Y: 1},   // significant deviation
				{X: 4, Y: 0},
			},
			tolerance: 0.05,
			expected: 3, // Should keep points with significant deviation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simplifyRoute(tt.points, tt.tolerance)
			if len(result) != tt.expected {
				t.Errorf("simplifyRoute() = %d points, expected %d points", len(result), tt.expected)
			}
		})
	}
}

func TestPerpendicularDistance(t *testing.T) {
	tests := []struct {
		name     string
		point    Point
		lineStart Point
		lineEnd   Point
		expected float64
		tolerance float64
	}{
		{
			name: "Point on line",
			point: Point{X: 0.5, Y: 0.5},
			lineStart: Point{X: 0, Y: 0},
			lineEnd: Point{X: 1, Y: 1},
			expected: 0,
			tolerance: 0.001,
		},
		{
			name: "Point above line",
			point: Point{X: 0.5, Y: 1},
			lineStart: Point{X: 0, Y: 0},
			lineEnd: Point{X: 1, Y: 1},
			expected: 0.353, // approximately sqrt(2)/4
			tolerance: 0.01,
		},
		{
			name: "Point below line",
			point: Point{X: 0.5, Y: 0},
			lineStart: Point{X: 0, Y: 0},
			lineEnd: Point{X: 1, Y: 1},
			expected: 0.353, // approximately sqrt(2)/4
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := perpendicularDistance(tt.point, tt.lineStart, tt.lineEnd)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("perpendicularDistance() = %f, expected %f (±%f)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		p1       Point
		p2       Point
		expected float64
		tolerance float64
	}{
		{
			name: "Same point",
			p1: Point{X: 0, Y: 0},
			p2: Point{X: 0, Y: 0},
			expected: 0,
			tolerance: 0.001,
		},
		{
			name: "Unit distance horizontal",
			p1: Point{X: 0, Y: 0},
			p2: Point{X: 1, Y: 0},
			expected: 1,
			tolerance: 0.001,
		},
		{
			name: "Unit distance vertical",
			p1: Point{X: 0, Y: 0},
			p2: Point{X: 0, Y: 1},
			expected: 1,
			tolerance: 0.001,
		},
		{
			name: "Diagonal distance",
			p1: Point{X: 0, Y: 0},
			p2: Point{X: 3, Y: 4},
			expected: 5, // 3-4-5 triangle
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distance(tt.p1, tt.p2)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("distance() = %f, expected %f (±%f)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test default configuration loading
	config := LoadConfig()
	
	// Test MQTT defaults
	if config.MQTT.Broker != "localhost" {
		t.Errorf("Expected MQTT broker 'localhost', got '%s'", config.MQTT.Broker)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected MQTT port 1883, got %d", config.MQTT.Port)
	}
	if config.MQTT.ClientID != "go_data_ingestion_client" {
		t.Errorf("Expected MQTT client ID 'go_data_ingestion_client', got '%s'", config.MQTT.ClientID)
	}
	if config.MQTT.Topic != "drivers_location/#" {
		t.Errorf("Expected MQTT topic 'drivers_location/#', got '%s'", config.MQTT.Topic)
	}

	// Test Redis defaults
	if config.Redis.Address != "127.0.0.1:6379" {
		t.Errorf("Expected Redis address '127.0.0.1:6379', got '%s'", config.Redis.Address)
	}
	if config.Redis.Password != "" {
		t.Errorf("Expected empty Redis password, got '%s'", config.Redis.Password)
	}
	if config.Redis.DB != 0 {
		t.Errorf("Expected Redis DB 0, got %d", config.Redis.DB)
	}

	// Test MongoDB defaults
	expectedMongoURI := "mongodb://root:examplepassword@127.0.0.1:27017"
	if config.MongoDB.URI != expectedMongoURI {
		t.Errorf("Expected MongoDB URI '%s', got '%s'", expectedMongoURI, config.MongoDB.URI)
	}
	if config.MongoDB.Database != "distributed_gps_route_tracking_system" {
		t.Errorf("Expected MongoDB database 'distributed_gps_route_tracking_system', got '%s'", config.MongoDB.Database)
	}
	if config.MongoDB.Collection != "trips" {
		t.Errorf("Expected MongoDB collection 'trips', got '%s'", config.MongoDB.Collection)
	}

	// Test route simplification defaults
	if config.RouteSimplification.Tolerance != 0.0001 {
		t.Errorf("Expected route tolerance 0.0001, got %f", config.RouteSimplification.Tolerance)
	}
}

// Benchmark for route simplification algorithm
func BenchmarkSimplifyRoute(b *testing.B) {
	// Create a large route for benchmarking
	points := make([]Point, 1000)
	for i := 0; i < 1000; i++ {
		// Create a sine wave pattern
		x := float64(i) / 10.0
		y := math.Sin(x) + (float64(i%10)-5)/100.0 // Add some noise
		points[i] = Point{X: x, Y: y}
	}

	tolerance := 0.01

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simplifyRoute(points, tolerance)
	}
}

// Benchmark for perpendicular distance calculation
func BenchmarkPerpendicularDistance(b *testing.B) {
	point := Point{X: 0.5, Y: 1}
	lineStart := Point{X: 0, Y: 0}
	lineEnd := Point{X: 1, Y: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		perpendicularDistance(point, lineStart, lineEnd)
	}
} 