package pkg

import (
	"github.com/go-spatial/geom"
	"reflect"
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

func TestProcessFeatures(t *testing.T) {
	var tests = []struct {
		geom          geom.Geometry
		expectedGeom  geom.Geometry
		resolution    float64
		replaceToggle bool
	}{
		// Lower single polygon resolution
		0: {geom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}}, expectedGeom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}}, resolution: float64(1), replaceToggle: false},
		// Higher single polygon resolution
		1: {geom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}}, expectedGeom: geom.MultiPolygon{{{{5, 5}}}}, resolution: float64(101), replaceToggle: true},
		// 2 hits on multi polygon
		2: {geom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{15, 15}, {15, 20}, {20, 20}, {20, 15}, {15, 15}}}}, expectedGeom: geom.MultiPolygon{{{{5, 5}}}, {{{17.5, 17.5}}}}, resolution: float64(101), replaceToggle: true},
		// 1 hit on multi polygon
		3: {geom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{15, 15}, {15, 20}, {20, 20}, {20, 15}, {15, 15}}}}, expectedGeom: geom.MultiPolygon{{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, {{{17.5, 17.5}}}}, resolution: float64(7), replaceToggle: true},
		// Polygon
		4: {geom: geom.Polygon{{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}}, expectedGeom: geom.Polygon{{{5, 5}}}, resolution: float64(101), replaceToggle: true},
	}

	for k, test := range tests {
		readyToWrite := make(chan Feature)
		needProcessing := make(chan Feature)

		go processFeatures(needProcessing, readyToWrite, test.resolution, test.replaceToggle)
		go testRoutineProcessFeature(readyToWrite, test.expectedGeom, t, k)

		testFeature := testFeature{
			geometry: test.geom,
		}
		needProcessing <- &testFeature

		close(needProcessing)
	}
}

func testRoutineProcessFeature(readyToWrite chan Feature, expectedGeom geom.Geometry, t *testing.T, k int) {
	for {
		feature, hasMore := <-readyToWrite
		if !hasMore {
			break
		} else {
			switch feature.Geometry().(type) {
			case geom.Polygon:
				p := feature.Geometry().(geom.Polygon).LinearRings()
				e := expectedGeom.(geom.Polygon).LinearRings()
				if !reflect.DeepEqual(p, e) {
					t.Errorf("test: %d, expected: %v \ngot: %v", k, e, p)
				}
			case geom.MultiPolygon:
				mp := feature.Geometry().(geom.MultiPolygon).Polygons()
				e := expectedGeom.(geom.MultiPolygon).Polygons()
				if !reflect.DeepEqual(mp, e) {
					t.Errorf("test: %d, expected: %v \ngot: %v", k, e, mp)
				}
			}
		}
	}
}

type testFeature struct {
	columns  []interface{}
	geometry geom.Geometry
}

func (f testFeature) Columns() []interface{} {
	return f.columns
}

func (f testFeature) Geometry() geom.Geometry {
	return f.geometry
}

func (f *testFeature) UpdateGeometry(geometry geom.Geometry) {
	f.geometry = geometry
}
