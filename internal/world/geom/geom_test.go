package geom

import (
	"math"
	"testing"
)

func TestDistanceAndClosestPoint(t *testing.T) {
	point := Vector2{X: 3, Y: 4}
	if got := Distance(Vector2{}, point); got != 5 {
		t.Fatalf("distance = %f, want 5", got)
	}
	closest := ClosestPointOnSegment(Vector2{X: 5, Y: 5}, Vector2{}, Vector2{X: 10, Y: 0})
	if closest != (Vector2{X: 5, Y: 0}) {
		t.Fatalf("closest point = %+v, want 5,0", closest)
	}
}

func TestProjectPoint(t *testing.T) {
	along, perpendicular := ProjectPoint(Vector2{}, Vector2{X: 1, Y: 0}, Vector2{X: 10, Y: 3})
	if along != 10 || perpendicular != 3 {
		t.Fatalf("project point = %f/%f, want 10/3", along, perpendicular)
	}
}

func TestSegmentsIntersect(t *testing.T) {
	if !SegmentsIntersect(Vector2{}, Vector2{X: 10, Y: 0}, Vector2{X: 5, Y: -5}, Vector2{X: 5, Y: 5}) {
		t.Fatal("segments should intersect")
	}
	if SegmentsIntersect(Vector2{}, Vector2{X: 10, Y: 0}, Vector2{X: 11, Y: -5}, Vector2{X: 11, Y: 5}) {
		t.Fatal("segments should not intersect")
	}
}

func TestSegmentEndpoints(t *testing.T) {
	start, end := SegmentEndpoints(Vector2{X: 10, Y: 10}, Vector2{X: 0, Y: 1}, 6)
	if math.Abs(start.Y-7) > 0.000001 || math.Abs(end.Y-13) > 0.000001 {
		t.Fatalf("endpoints = %+v/%+v, want y 7/13", start, end)
	}
}
