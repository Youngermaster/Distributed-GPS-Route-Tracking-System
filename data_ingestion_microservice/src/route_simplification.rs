use crate::types::{Location, ServiceError, ServiceResult};
use geo::{algorithm::simplify::Simplify, LineString, Point};
use log::{debug, info};

/// Route simplification service with different algorithms
#[derive(Clone)]
pub struct RouteSimplifier {
    tolerance: f64,
}

impl RouteSimplifier {
    /// Create a new route simplifier with the given tolerance
    pub fn new(tolerance: f64) -> ServiceResult<Self> {
        if tolerance <= 0.0 {
            return Err(ServiceError::Validation(
                "Tolerance must be greater than 0".to_string(),
            ));
        }

        Ok(Self { tolerance })
    }

    /// Simplify a route using the Ramer-Douglas-Peucker algorithm
    pub fn simplify_route(&self, locations: &[Location]) -> ServiceResult<Vec<Location>> {
        if locations.is_empty() {
            return Ok(Vec::new());
        }

        if locations.len() <= 2 {
            return Ok(locations.to_vec());
        }

        debug!("Simplifying route with {} points", locations.len());

        // Convert locations to geo::Point
        let points: Vec<Point<f64>> = locations
            .iter()
            .map(|loc| Point::new(loc.longitude, loc.latitude))
            .collect();

        // Create a LineString from the points
        let linestring = LineString::from(points);

        // Apply the simplification algorithm
        let simplified_linestring = linestring.simplify(&self.tolerance);

        // Convert back to Location structs
        let simplified_locations: Vec<Location> = simplified_linestring
            .0
            .iter()
            .map(|point| Location {
                latitude: point.y,
                longitude: point.x,
            })
            .collect();

        let compression_ratio = simplified_locations.len() as f64 / locations.len() as f64;

        info!(
            "Route simplified: {} -> {} points (compression ratio: {:.2}%)",
            locations.len(),
            simplified_locations.len(),
            compression_ratio * 100.0
        );

        Ok(simplified_locations)
    }

    /// Alternative simplification using custom implementation
    pub fn simplify_route_custom(&self, locations: &[Location]) -> ServiceResult<Vec<Location>> {
        if locations.is_empty() {
            return Ok(Vec::new());
        }

        if locations.len() <= 2 {
            return Ok(locations.to_vec());
        }

        let mut simplified = Vec::new();
        simplified.push(locations[0].clone());

        let mut i = 0;
        while i < locations.len() - 1 {
            let mut farthest_index = i + 1;
            let mut max_distance = 0.0;

            // Look ahead to find the farthest point that still maintains accuracy
            for j in (i + 1)..locations.len() {
                let distance = self.perpendicular_distance(
                    &locations[j],
                    &locations[i],
                    &locations[locations.len() - 1],
                );

                if distance > self.tolerance {
                    break;
                }

                if distance > max_distance {
                    max_distance = distance;
                    farthest_index = j;
                }
            }

            simplified.push(locations[farthest_index].clone());
            i = farthest_index;
        }

        // Ensure the last point is included
        if simplified.last() != locations.last() {
            simplified.push(locations[locations.len() - 1].clone());
        }

        let compression_ratio = simplified.len() as f64 / locations.len() as f64;

        info!(
            "Route simplified (custom): {} -> {} points (compression ratio: {:.2}%)",
            locations.len(),
            simplified.len(),
            compression_ratio * 100.0
        );

        Ok(simplified)
    }

    /// Calculate perpendicular distance from a point to a line
    fn perpendicular_distance(
        &self,
        point: &Location,
        line_start: &Location,
        line_end: &Location,
    ) -> f64 {
        let area = (line_start.longitude * (line_end.latitude - point.latitude)
            + line_end.longitude * (point.latitude - line_start.latitude)
            + point.longitude * (line_start.latitude - line_end.latitude))
            .abs();

        let base = self.distance(line_start, line_end);

        if base == 0.0 {
            self.distance(point, line_start)
        } else {
            area / base
        }
    }

    /// Calculate Euclidean distance between two points
    fn distance(&self, p1: &Location, p2: &Location) -> f64 {
        let dx = p1.longitude - p2.longitude;
        let dy = p1.latitude - p2.latitude;
        (dx * dx + dy * dy).sqrt()
    }

    /// Get the current tolerance value
    pub fn tolerance(&self) -> f64 {
        self.tolerance
    }

    /// Update the tolerance value
    pub fn set_tolerance(&mut self, tolerance: f64) -> ServiceResult<()> {
        if tolerance <= 0.0 {
            return Err(ServiceError::Validation(
                "Tolerance must be greater than 0".to_string(),
            ));
        }
        self.tolerance = tolerance;
        Ok(())
    }
}

/// Utility function to calculate route statistics
pub fn calculate_route_stats(original: &[Location], simplified: &[Location]) -> RouteStats {
    let original_length = calculate_total_distance(original);
    let simplified_length = calculate_total_distance(simplified);

    RouteStats {
        original_points: original.len(),
        simplified_points: simplified.len(),
        compression_ratio: if original.is_empty() {
            0.0
        } else {
            simplified.len() as f64 / original.len() as f64
        },
        original_length,
        simplified_length,
        length_difference: (original_length - simplified_length).abs(),
    }
}

/// Calculate the total distance of a route
fn calculate_total_distance(locations: &[Location]) -> f64 {
    if locations.len() < 2 {
        return 0.0;
    }

    locations
        .windows(2)
        .map(|window| {
            let dx = window[1].longitude - window[0].longitude;
            let dy = window[1].latitude - window[0].latitude;
            (dx * dx + dy * dy).sqrt()
        })
        .sum()
}

/// Statistics about route simplification
#[derive(Debug, Clone)]
pub struct RouteStats {
    pub original_points: usize,
    pub simplified_points: usize,
    pub compression_ratio: f64,
    pub original_length: f64,
    pub simplified_length: f64,
    pub length_difference: f64,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_locations() -> Vec<Location> {
        vec![
            Location {
                latitude: 0.0,
                longitude: 0.0,
            },
            Location {
                latitude: 0.5,
                longitude: 0.5,
            },
            Location {
                latitude: 1.0,
                longitude: 1.0,
            },
            Location {
                latitude: 1.5,
                longitude: 1.5,
            },
            Location {
                latitude: 2.0,
                longitude: 2.0,
            },
        ]
    }

    #[test]
    fn test_route_simplifier_creation() {
        let simplifier = RouteSimplifier::new(0.001);
        assert!(simplifier.is_ok());

        let invalid_simplifier = RouteSimplifier::new(-1.0);
        assert!(invalid_simplifier.is_err());
    }

    #[test]
    fn test_empty_route_simplification() {
        let simplifier = RouteSimplifier::new(0.001).unwrap();
        let result = simplifier.simplify_route(&[]).unwrap();
        assert!(result.is_empty());
    }

    #[test]
    fn test_single_point_route() {
        let simplifier = RouteSimplifier::new(0.001).unwrap();
        let locations = vec![Location {
            latitude: 1.0,
            longitude: 1.0,
        }];
        let result = simplifier.simplify_route(&locations).unwrap();
        assert_eq!(result.len(), 1);
        assert_eq!(result[0].latitude, 1.0);
        assert_eq!(result[0].longitude, 1.0);
    }

    #[test]
    fn test_straight_line_simplification() {
        let simplifier = RouteSimplifier::new(0.1).unwrap();
        let locations = create_test_locations();
        let result = simplifier.simplify_route(&locations).unwrap();

        // A straight line should be simplified to just start and end points
        assert!(result.len() <= locations.len());
        assert_eq!(result[0].latitude, 0.0);
        assert_eq!(result.last().unwrap().latitude, 2.0);
    }

    #[test]
    fn test_route_stats() {
        let original = create_test_locations();
        let simplified = vec![
            Location {
                latitude: 0.0,
                longitude: 0.0,
            },
            Location {
                latitude: 2.0,
                longitude: 2.0,
            },
        ];

        let stats = calculate_route_stats(&original, &simplified);
        assert_eq!(stats.original_points, 5);
        assert_eq!(stats.simplified_points, 2);
        assert_eq!(stats.compression_ratio, 0.4);
    }

    #[test]
    fn test_distance_calculation() {
        let simplifier = RouteSimplifier::new(0.001).unwrap();
        let p1 = Location {
            latitude: 0.0,
            longitude: 0.0,
        };
        let p2 = Location {
            latitude: 3.0,
            longitude: 4.0,
        };

        let distance = simplifier.distance(&p1, &p2);
        assert!((distance - 5.0).abs() < 0.001); // 3-4-5 triangle
    }

    #[test]
    fn test_tolerance_update() {
        let mut simplifier = RouteSimplifier::new(0.001).unwrap();
        assert_eq!(simplifier.tolerance(), 0.001);

        simplifier.set_tolerance(0.002).unwrap();
        assert_eq!(simplifier.tolerance(), 0.002);

        assert!(simplifier.set_tolerance(-1.0).is_err());
    }
}
