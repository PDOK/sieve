package main

import "github.com/go-spatial/geom"

type feature interface {
	Columns() []interface{}
	Geometry() geom.Geometry
	GeometrySize() int
	UpdateGeometry(geom.Geometry)
}

type features []interface{}

// Source handles state and reads it to from a DataSource
type Source interface {
	ReadFeatures(chan feature, chan int)
}

// Target handles state and writes it to a DataTarget
type Target interface {
	WriteFeatures(features)
	PageSize() int
	MemoryLimit() (float64, bool)
}
