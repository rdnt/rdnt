package main

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	svg "github.com/ajstarks/svgo/float"
	"github.com/akrennmair/slice"

	"readme/pkg/github"
)

type Col struct {
	pts   []Vector3
	count int
}

type Vector2 struct {
	X float64
	Y float64
}

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	contribs, err := github.Contributions(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	var cols []Col

	for k, c := range contribs {
		z := float64(c.Count*2 + 2)

		i := float64(k / 7)
		j := float64(k % 7)

		var pts []Vector3

		// top
		pts = append(pts, Vector3{X: 12 * i, Y: 12 * j, Z: -z})
		pts = append(pts, Vector3{X: 12*i + 10, Y: 12 * j, Z: -z})
		pts = append(pts, Vector3{X: 12*i + 10, Y: 12*j + 10, Z: -z})
		pts = append(pts, Vector3{X: 12 * i, Y: 12*j + 10, Z: -z})

		// bottom
		pts = append(pts, Vector3{X: 12 * i, Y: 12 * j, Z: 0})
		pts = append(pts, Vector3{X: 12*i + 10, Y: 12 * j, Z: 0})
		pts = append(pts, Vector3{X: 12*i + 10, Y: 12*j + 10, Z: 0})
		pts = append(pts, Vector3{X: 12 * i, Y: 12*j + 10, Z: 0})

		cols = append(cols, Col{pts: pts, count: c.Count})
	}

	canvas := svg.New(os.Stdout)
	// approximate size should do for now
	canvas.Start(880, 450)

	for _, c := range cols {
		// each column has 3 visible faces.
		// p1 is the outline of the column (all 3 visible faces), p2 is the bottom 2 faces and 3 is the bottom right face
		// they are rendered that way to avoid weird spacing artifacts between the faces
		p1 := []Vector3{c.pts[0], c.pts[1], c.pts[5], c.pts[6], c.pts[7], c.pts[3]}
		p2 := []Vector3{c.pts[3], c.pts[2], c.pts[1], c.pts[5], c.pts[6], c.pts[7]}
		p3 := []Vector3{c.pts[2], c.pts[1], c.pts[5], c.pts[6]}

		// project the 3d coordinates to 2d space using an isometric projection
		p1iso := isometricProjection(p1)
		p2iso := isometricProjection(p2)
		p3iso := isometricProjection(p3)

		// horizontal & vertical offsets to center the whole chart
		h, v := float64(200), float64(50)

		// colors used for the 3 visible faces
		c1, c2, c3 := faceColors(c.count)

		xs := slice.Map(p1iso, func(vec Vector2) float64 { return vec.X + h })
		ys := slice.Map(p1iso, func(vec Vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill:"+c1)

		xs = slice.Map(p2iso, func(vec Vector2) float64 { return vec.X + h })
		ys = slice.Map(p2iso, func(vec Vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill:"+c2)

		xs = slice.Map(p3iso, func(vec Vector2) float64 { return vec.X + h })
		ys = slice.Map(p3iso, func(vec Vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill:"+c3)
	}

	// save the svg
	canvas.End()
}

func isometricProjection(v []Vector3) []Vector2 {
	return slice.Map(v, func(v Vector3) Vector2 {
		x, y := spaceToIso(v.X, v.Y, v.Z)

		return Vector2{
			X: x,
			Y: y,
		}
	})
}

func spaceToIso(x, y, z float64) (h, v float64) {
	x, y = x+z, y+z

	h = (x - y) * math.Sqrt(3) / 2
	v = (x + y) / 2

	return h, v
}

func faceColors(r int) (string, string, string) {
	if r > 10 {
		return "#39d353", "#10a92c", "#24bd40"
	} else if r > 7 {
		return "#26a641", "#007d1a", "#11912e"
	} else if r > 3 {
		return "#006d32", "#004307", "#00571b"
	} else if r > 0 {
		return "#0e4429", "#001b00", "#002f12"
	} else {
		return "#2d333b", "#030a12", "#171e26"
	}
}
