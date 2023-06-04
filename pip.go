package pip

import (
	"runtime"
	"sync"

	"golang.org/x/exp/constraints"
)

// type Number = constraints.Number
type Float = constraints.Float

type Point[T Float] struct {
	// A point
	X T
	Y T
}

type Point64 = Point[float64]
type Point32 = Point[float32]

type Polygon[T Float] struct {
	// A polygon
	Points []Point[T]
}

type Polygon64 = Polygon[float64]

type BoundingBox[T Float] struct {
	BottomLeft Point[T]
	TopRight   Point[T]
}

func PointInPolygon[T Float](pt Point[T], poly Polygon[T]) bool {
	// Checks if point is inside polygon

	bb := GetBoundingBox(poly) // Get the bounding box of the polygon in question

	// If point not in bounding box return false immediately
	if !PointInBoundingBox(pt, bb) {
		return false
	}

	// If the point is in the bounding box then we need to check the polygon
	nverts := len(poly.Points)
	intersect := false

	verts := poly.Points
	j := 0

	for i := 1; i < nverts; i++ {

		if ((verts[i].Y > pt.Y) != (verts[j].Y > pt.Y)) &&
			(pt.X < (verts[j].X-verts[i].X)*(pt.Y-verts[i].Y)/(verts[j].Y-verts[i].Y)+verts[i].X) {
			intersect = !intersect
		}

		j = i

	}

	return intersect

}

func MaxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

func PointInPolygonParallel[T Float](pts []Point[T], poly Polygon[T], numcores int) []Point[T] {

	MAXPROCS := MaxParallelism()
	runtime.GOMAXPROCS(MAXPROCS)

	if numcores > MAXPROCS {
		numcores = MAXPROCS
	}

	start := 0
	inside := []Point[T]{}

	var m sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numcores)

	for i := 1; i <= numcores; i++ {

		size := (len(pts) / numcores) * i
		batch := pts[start:size]

		go func(batch []Point[T]) {
			defer wg.Done()
			for j := 0; j < len(batch); j++ {
				pt := batch[j]
				if PointInPolygon(pt, poly) {
					m.Lock()
					inside = append(inside, pt)
					m.Unlock()
				}
			}

		}(batch)

		start = size
	}

	wg.Wait()

	return inside

}

func PointInBoundingBox[T Float](pt Point[T], bb BoundingBox[T]) bool {
	// Check if point is in bounding box

	// Bottom Left is the smallest and x and y value
	// Top Right is the largest x and y value
	return pt.X < bb.TopRight.X && pt.X > bb.BottomLeft.X &&
		pt.Y < bb.TopRight.Y && pt.Y > bb.BottomLeft.Y

}

func GetBoundingBox[T Float](poly Polygon[T]) BoundingBox[T] {

	var maxX, maxY, minX, minY T

	for i := 0; i < len(poly.Points); i++ {
		side := poly.Points[i]

		if side.X > maxX || maxX == 0.0 {
			maxX = side.X
		}
		if side.Y > maxY || maxY == 0.0 {
			maxY = side.Y
		}
		if side.X < minX || minX == 0.0 {
			minX = side.X
		}
		if side.Y < minY || minY == 0.0 {
			minY = side.Y
		}
	}

	return BoundingBox[T]{
		BottomLeft: Point[T]{X: minX, Y: minY},
		TopRight:   Point[T]{X: maxX, Y: maxY},
	}

}
