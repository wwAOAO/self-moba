package geom

import (
	"l-battle/internal/world/formula"
	"math"
)

type Vector2 struct {
	X float64
	Y float64
}

func Distance(a Vector2, b Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func DistancePointToSegment(point Vector2, start Vector2, end Vector2) float64 {
	return Distance(point, ClosestPointOnSegment(point, start, end))
}

func ClosestPointOnSegment(point Vector2, start Vector2, end Vector2) Vector2 {
	dx := end.X - start.X
	dy := end.Y - start.Y
	lengthSquared := dx*dx + dy*dy
	if lengthSquared <= 0 {
		return start
	}
	t := ((point.X-start.X)*dx + (point.Y-start.Y)*dy) / lengthSquared
	t = formula.Clamp(t, 0, 1)
	return Vector2{
		X: start.X + dx*t,
		Y: start.Y + dy*t,
	}
}

func ProjectPoint(origin Vector2, direction Vector2, point Vector2) (float64, float64) {
	dx := point.X - origin.X
	dy := point.Y - origin.Y
	along := dx*direction.X + dy*direction.Y
	perpX := dx - along*direction.X
	perpY := dy - along*direction.Y
	return along, math.Hypot(perpX, perpY)
}

func SegmentEndpoints(center Vector2, direction Vector2, width float64) (Vector2, Vector2) {
	half := width / 2
	return Vector2{
			X: center.X - direction.X*half,
			Y: center.Y - direction.Y*half,
		}, Vector2{
			X: center.X + direction.X*half,
			Y: center.Y + direction.Y*half,
		}
}

func SegmentsIntersect(a Vector2, b Vector2, c Vector2, d Vector2) bool {
	ab1 := Orientation(a, b, c)
	ab2 := Orientation(a, b, d)
	cd1 := Orientation(c, d, a)
	cd2 := Orientation(c, d, b)
	return ab1*ab2 <= 0 && cd1*cd2 <= 0
}

func Orientation(a Vector2, b Vector2, c Vector2) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}
