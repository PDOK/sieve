package pkg

import (
	"testing"
)

func TestShoelace(t *testing.T) {
	var tests = []struct {
		pts  [][2]float64
		area float64
	}{
		// Rectangle
		0: {pts: [][2]float64{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, area: float64(100)},
		// Triangle
		1: {pts: [][2]float64{{0, 0}, {5, 10}, {0, 10}, {0, 0}}, area: float64(25)},
		// Missing 'official closing point
		2: {pts: [][2]float64{{0, 0}, {0, 10}, {10, 10}, {10, 0}}, area: float64(100)},
		// Single point
		3: {pts: [][2]float64{{1234, 4321}}, area: float64(0.000000)},
		// No point
		4: {pts: nil, area: float64(0.000000)},
		// Empty point
		5: {pts: [][2]float64{}, area: float64(0.000000)},
	}

	for k, test := range tests {
		area := shoelace(test.pts)
		if area != test.area {
			t.Errorf("test: %d, expected: %f \ngot: %f", k, test.area, area)
		}
	}
}

func TestArea(t *testing.T) {
	var tests = []struct {
		geom [][][2]float64
		area float64
	}{
		// Rectangle
		0: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, area: float64(100)},
		// Rectangle with hole
		1: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {{2, 2}, {2, 8}, {8, 8}, {8, 2}, {2, 2}}}, area: float64(64)},
		// Rectangle with empty hole
		2: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {}}, area: float64(100)},
		// Rectangle with nil hole
		3: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, nil}, area: float64(100)},
		// nil geometry
		4: {geom: nil, area: float64(0)},
	}

	for k, test := range tests {
		area := area(test.geom)
		if area != test.area {
			t.Errorf("test: %d, expected: %f \ngot: %f", k, test.area, area)
		}
	}
}

func TestPolygonCentroid(t *testing.T) {
	var tests = []struct {
		geom     [][][2]float64
		expected [2]float64
	}{
		// Rectangle
		0: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, expected: [2]float64{5, 5}},
		// Rectangle with hole
		1: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {{2, 2}, {2, 8}, {8, 8}, {8, 2}, {2, 2}}}, expected: [2]float64{5, 5}},
		// Rectangle with empty hole
		2: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {}}, expected: [2]float64{5, 5}},
		// Rectangle with nil hole
		3: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, nil}, expected: [2]float64{5, 5}},
		// nil geometry
		4: {geom: nil, expected: [2]float64{0, 0}},
	}

	for k, test := range tests {
		result := getPolygonCentroid(test.geom)
		if result != test.expected {
			t.Errorf("test: %d, expected: %f \ngot: %f", k, test.expected, result)
		}
	}
}

func TestPolygonSieve(t *testing.T) {
	var tests = []struct {
		geom           [][][2]float64
		resolution     float64
		expectedSieved bool
	}{
		// Lower resolution
		0: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, resolution: float64(9), expectedSieved: false},
		// Higher resolution
		1: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, resolution: float64(101), expectedSieved: true},
		// Nil input
		2: {geom: nil, resolution: float64(1), expectedSieved: false},
		// Filterout donut
		3: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {{5, 5}, {5, 6}, {6, 6}, {6, 5}, {5, 5}}}, resolution: float64(9), expectedSieved: false},
		// Donut stays
		4: {geom: [][][2]float64{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}, {{5, 5}, {5, 8.5}, {8.5, 8.5}, {8.5, 5}, {5, 5}}}, resolution: float64(3), expectedSieved: false},
	}

	for k, test := range tests {
		_, sieved := polygonSieve(test.geom, test.resolution)

		if test.expectedSieved != sieved {
			t.Errorf("test: %d, expected: %t \ngot: %t", k, test.expectedSieved, sieved)
		}
	}
}

func TestMultiPolygonSieve(t *testing.T) {
	var tests = []struct {
		geom           [][][][2]float64
		resolution     float64
		expectedSieved bool
	}{
		// Lower single polygon resolution
		0: {geom: [][][][2]float64{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}}, resolution: float64(1), expectedSieved: false},
		// Higher single polygon resolution
		1: {geom: [][][][2]float64{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}}, resolution: float64(101), expectedSieved: true},
		// Nil input
		2: {geom: nil, resolution: float64(1), expectedSieved: false},
		// Low multi polygon resolution
		3: {geom: [][][][2]float64{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{15, 15}, {15, 20}, {20, 20}, {20, 15}, {15, 15}}}}, resolution: float64(1), expectedSieved: false},
		// single hit on multi polygon
		4: {geom: [][][][2]float64{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{15, 15}, {15, 20}, {20, 20}, {20, 15}, {15, 15}}}}, resolution: float64(9), expectedSieved: true},
		// single hit on multi polygon
		5: {geom: [][][][2]float64{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{15, 15}, {15, 20}, {20, 20}, {20, 15}, {15, 15}}}}, resolution: float64(101), expectedSieved: true},
	}

	for k, test := range tests {
		_, sieved := multiPolygonSieve(test.geom, test.resolution)

		if test.expectedSieved != sieved {
			t.Errorf("test: %d, expected: %t \ngot: %t", k, test.expectedSieved, sieved)
		}
	}
}
