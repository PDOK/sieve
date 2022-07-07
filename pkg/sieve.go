package pkg

import (
	geometry "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
	"log"
	"math"

	"github.com/go-spatial/geom"
)

// readFeatures reads the features from the given Geopackage table
// and decodes the WKB geometry to a geom.Polygon
func readFeaturesFromSource(source Source, preSieve chan Feature) {
	source.ReadFeatures(preSieve)
}

// sieveFeatures sieves/filters the geometry against the given resolution
// the two steps that are done are:
// 1. filter features with a area smaller then the (resolution*resolution)
// 2. removes interior rings with a area smaller then the (resolution*resolution)
func sieveFeatures(preSieve chan Feature, postSieve chan Feature, resolution float64) {
	var preSieveCount, postSieveCount, nonPolygonCount, multiPolygonCount uint64
	for {
		feature, hasMore := <-preSieve
		if !hasMore {
			break
		} else {
			preSieveCount++
			switch feature.Geometry().(type) {
			case geom.Polygon:
				p := feature.Geometry().(geom.Polygon)
				sievedPolygon, isReduced := polygonSieve(p, resolution)
				feature.UpdateGeometry(sievedPolygon)
				feature.IsReduced(isReduced)
				if !isReduced {
					postSieveCount++
				}
				postSieve <- feature
			case geom.MultiPolygon:
				mp := feature.Geometry().(geom.MultiPolygon)
				mp, isReduced := multiPolygonSieve(mp, resolution)
				feature.UpdateGeometry(mp)
				feature.IsReduced(isReduced)
				multiPolygonCount++
				if !isReduced {
					postSieveCount++
				}
				postSieve <- feature
			default:
				postSieveCount++
				nonPolygonCount++
				feature.IsReduced(false)
				postSieve <- feature
			}
		}
	}
	close(postSieve)

	log.Printf("    total features: %d", preSieveCount)
	log.Printf("      non-polygons: %d", nonPolygonCount)
	if preSieveCount != nonPolygonCount {
		log.Printf("     multipolygons: %d", multiPolygonCount)
	}
	log.Printf("       not reduced: %d", postSieveCount)
}

// writeFeatures collects the processed features by the sieveFeatures and
// creates a WKB binary from the geometry
// The collected feature array, based on the pagesize, is then passed to the writeFeaturesArray
func writeFeaturesToTarget(postSieve chan Feature, kill chan bool, target Target) {

	target.WriteFeatures(postSieve)
	kill <- true
}

// getMultiPolygonCentroid returns Point with the centroid value of a multiPolygon
func getMultiPolygonCentroid(mp geom.MultiPolygon) geom.Point {
	var multiPolygonCoords [][][]geometry.Coord
	for _, p := range mp {
		multiPolygonCoords = append(multiPolygonCoords, getPolygonCoords(p))
	}
	if len(multiPolygonCoords) != 0 {
		centroidCoord, err := xy.Centroid(geometry.NewMultiPolygon(geometry.XY).MustSetCoords(multiPolygonCoords))
		if err != nil {
			panic(err)
		}
		return [2]float64{centroidCoord[0], centroidCoord[1]}
	} else {
		return [2]float64{0, 0}
	}
}

// getPolygonCentroid returns Point with the centroid value of a polygon
func getPolygonCentroid(p geom.Polygon) geom.Point {
	polygonCoords := getPolygonCoords(p)
	for _, test := range polygonCoords {
		if len(test) == 3 {
			print("oei?")
		}
	}
	if polygonCoords != nil {
		polygon := geometry.NewPolygon(geometry.XY).MustSetCoords(polygonCoords)
		centroidCoord, err := xy.Centroid(polygon)
		if err != nil {
			panic(err)
		}
		return [2]float64{centroidCoord[0], centroidCoord[1]}
	} else {
		return [2]float64{0, 0}
	}
}

// multiPolygonSieve will split it self into the separated polygons that will be sieved before building a new MULTIPOLYGON
func multiPolygonSieve(mp geom.MultiPolygon, resolution float64) (geom.MultiPolygon, bool) {
	var sievedMultiPolygon geom.MultiPolygon
	var isReduced bool
	for _, p := range mp {
		if sievedPolygon, b := polygonSieve(p, resolution); b == false {
			sievedMultiPolygon = append(sievedMultiPolygon, sievedPolygon)
		} else {
			isReduced = true
			sievedMultiPolygon = append(sievedMultiPolygon, [][][2]float64{{getPolygonCentroid(p)}})
		}
	}
	return sievedMultiPolygon, isReduced
}

// polygonSieve will sieve a given POLYGON
func polygonSieve(p geom.Polygon, resolution float64) (geom.Polygon, bool) {
	minArea := resolution * resolution
	if area(p) > minArea {
		if len(p) > 1 {
			var sievedPolygon geom.Polygon
			sievedPolygon = append(sievedPolygon, p[0])
			for _, interior := range p[1:] {
				if shoelace(interior) > minArea {
					sievedPolygon = append(sievedPolygon, interior)
				}
			}
			return sievedPolygon, false
		}
		return p, false
	}
	if p == nil {
		return nil, false
	} else {
		return [][][2]float64{{getPolygonCentroid(p)}}, true
	}
}

// calculate the area of a polygon
func area(geom [][][2]float64) float64 {
	interior := .0
	if geom == nil {
		return 0.
	}
	if len(geom) > 1 {
		for _, i := range geom[1:] {
			interior = interior + shoelace(i)
		}
	}
	return shoelace(geom[0]) - interior
}

// https://en.wikipedia.org/wiki/Shoelace_formula
func shoelace(pts [][2]float64) float64 {
	sum := 0.
	if len(pts) == 0 {
		return 0.
	}

	p0 := pts[len(pts)-1]
	for _, p1 := range pts {
		sum += p0[1]*p1[0] - p0[0]*p1[1]
		p0 = p1
	}
	return math.Abs(sum / 2)
}

func Sieve(source Source, target Target, resolution float64) {

	preSieve := make(chan Feature)
	postSieve := make(chan Feature)
	kill := make(chan bool)

	go writeFeaturesToTarget(postSieve, kill, target)
	go sieveFeatures(preSieve, postSieve, resolution)
	go readFeaturesFromSource(source, preSieve)

	for {
		if <-kill {
			break
		}
	}
	close(kill)
}

func getPolygonCoords(p geom.Polygon) [][]geometry.Coord {
	var multiXyCoordinates [][]geometry.Coord
	if p == nil {
		return nil
	}
	for _, polygon := range p {
		var xyCoordinates []geometry.Coord
		if len(polygon) != 0 {
			for _, point := range polygon {
				xyCoordinates = append(xyCoordinates, makeCoord(point))
			}
			// Close the geometry
			xyCoordinates = append(xyCoordinates, xyCoordinates[0])
			multiXyCoordinates = append(multiXyCoordinates, xyCoordinates)
		}
	}
	return multiXyCoordinates
}

func makeCoord(point [2]float64) geometry.Coord {
	return geometry.Coord{point[0], point[1]}
}
