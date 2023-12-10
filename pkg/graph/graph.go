// Package graph allows rendering an isometric projection of a GitHub user's contribution history.
package graph

import (
	"fmt"
	"io"
	"math"

	svg "github.com/ajstarks/svgo/float"
	"github.com/akrennmair/slice"

	"github.com/rdnt/rdnt/pkg/github"
)

// Graph is a graph instance that is used to render a contributions graph.
type Graph struct {
	isHalloween  bool
	cols         []col
	total        int
	priv         int
	issues       int
	pullRequests int
}

// NewGraph creates a graph instance with the contributions-per-day that are passed. Graph.Render can then be used to
// render the graph.
func NewGraph(contribs github.Contributions, stats github.Stats) *Graph {
	var cols []col
	var total int

	for k, entry := range contribs.Contributions {
		total += entry.Count

		z := float64(entry.Count*2 + 2)

		i := float64(k / 7)
		j := float64(k % 7)

		size := float64(10)
		space := float64(12)

		var pts []vector3
		// top
		pts = append(pts, vector3{X: space * i, Y: space * j, Z: -z})
		pts = append(pts, vector3{X: space*i + size, Y: space * j, Z: -z})
		pts = append(pts, vector3{X: space*i + size, Y: space*j + size, Z: -z})
		pts = append(pts, vector3{X: space * i, Y: space*j + size, Z: -z})
		// bottom
		pts = append(pts, vector3{X: space * i, Y: space * j, Z: 0})
		pts = append(pts, vector3{X: space*i + size, Y: space * j, Z: 0})
		pts = append(pts, vector3{X: space*i + size, Y: space*j + size, Z: 0})
		pts = append(pts, vector3{X: space * i, Y: space*j + size, Z: 0})

		cols = append(cols, col{pts: pts, count: entry.Count, color: entry.Color})
	}

	return &Graph{
		cols:         cols,
		total:        stats.TotalContribs + stats.PrivateContribs,
		priv:         stats.PrivateContribs,
		isHalloween:  contribs.IsHalloween,
		issues:       stats.Issues,
		pullRequests: stats.PullRequests,
	}
}

// Render writes the generated SVG to the given io.Writer in the appropriate Theme.
func (g *Graph) Render(f io.WriteCloser, theme Theme) error {
	canvas := svg.New(f)

	// approximate size should do for now
	canvas.Start(840, 400)

	for _, c := range g.cols {
		// each column has 3 visible faces.
		// p1 is the outline of the column (all 3 visible faces), p2 is the bottom 2 faces and 3 is the bottom right face
		// they are rendered that way to avoid weird spacing artifacts between the faces
		p1 := []vector3{c.pts[0], c.pts[1], c.pts[5], c.pts[6], c.pts[7], c.pts[3]}
		p2 := []vector3{c.pts[3], c.pts[2], c.pts[1], c.pts[5], c.pts[6], c.pts[7]}
		p3 := []vector3{c.pts[2], c.pts[1], c.pts[5], c.pts[6]}

		// project the 3d coordinates to 2d space using an isometric projection
		p1iso := isometricProjection(p1)
		p2iso := isometricProjection(p2)
		p3iso := isometricProjection(p3)

		// horizontal & vertical offsets to center the whole chart
		h, v := float64(180), float64(25)

		// colors used for the 3 visible faces
		c1, c2, c3 := faceColors(c.color, theme, g.isHalloween)

		xs := slice.Map(p1iso, func(vec vector2) float64 { return vec.X + h })
		ys := slice.Map(p1iso, func(vec vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill: "+c1)

		xs = slice.Map(p2iso, func(vec vector2) float64 { return vec.X + h })
		ys = slice.Map(p2iso, func(vec vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill: "+c2)

		xs = slice.Map(p3iso, func(vec vector2) float64 { return vec.X + h })
		ys = slice.Map(p3iso, func(vec vector2) float64 { return vec.Y + v })
		canvas.Polygon(xs, ys, "fill: "+c3)
	}

	var c1, c2 string

	if g.isHalloween {
		c1, _, _ = faceColors(halloweenColor0, 1-theme, true)
		c2, _, _ = faceColors(halloweenColor1, 1-theme, true)
	} else {
		c1, _, _ = faceColors(color0, 1-theme, false)
		c2, _, _ = faceColors(color1, 1-theme, false)
	}

	deci := canvas.Decimals
	font := `-apple-system,BlinkMacSystemFont,'Segoe UI','Noto Sans',Helvetica,Arial,sans-serif,'Apple Color Emoji','Segoe UI Emoji'`

	canvas.Gstyle(fmt.Sprintf("text-anchor:%s;font-family:%s;", "end", font))

	stats := [][2]string{
		{"Pull Requests", fmt.Sprint(g.pullRequests)},
		{"Issues", fmt.Sprint(g.issues)},
	}

	for i, stat := range stats {
		canvas.Text(840-100-10, float64((i*72)+24)+10, stat[0], fmt.Sprintf("font-weight:400;fill:%s;font-size:%.*fpx;", c1, deci, 20.0))
		canvas.Text(840-100-10, float64((i*72)+56)+10, stat[1], fmt.Sprintf("font-weight:700;fill:%s;font-size:%.*fpx;", c2, deci, 24.0))
	}

	canvas.Gend()

	total := [2]string{"Total contributions", fmt.Sprint(g.total)}

	canvas.Gstyle(fmt.Sprintf("text-anchor:%s;font-family:%s;", "start", font))

	canvas.Text(100+10, 400-62.0-8-10, total[0], fmt.Sprintf("font-weight:400;fill:%s;font-size:%.*fpx;", c1, deci, 24.0))
	canvas.Text(100+10, 400-28.0-8-10, total[1], fmt.Sprintf("font-weight:700;fill:%s;font-size:%.*fpx;", c2, deci, 28.0))

	canvas.Gend()

	// save the svg
	canvas.End()

	return f.Close()
}

func isometricProjection(v []vector3) []vector2 {
	return slice.Map(v, func(v vector3) vector2 {
		x, y := spaceToIso(v.X, v.Y, v.Z)

		return vector2{
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

const (
	color0 = "#ebedf0"
	color1 = "#9be9a8"
	color2 = "#40c463"
	color3 = "#30a14e"
	color4 = "#216e39"

	halloweenColor0 = "#ebedf0"
	halloweenColor1 = "#ffee4a"
	halloweenColor2 = "#ffc501"
	halloweenColor3 = "#fe9600"
	halloweenColor4 = "#03001c"
)

func faceColors(color string, theme Theme, isHalloween bool) (string, string, string) {
	switch theme {
	case Dark:
		if isHalloween {
			switch color {
			case halloweenColor0:
				return "#161b22", "#000000", "#00050c"
			case halloweenColor1:
				return "#631c03", "#390000", "#4d0800"
			case halloweenColor2:
				return "#bd561d", "#942d00", "#a84109"
			case halloweenColor3:
				return "#fa7a18", "#d05000", "#e46403"
			case halloweenColor4:
				return "#fddf68", "#d3b53e", "#e7c952"
			}
		} else {
			switch color {
			case color0:
				return "#2d333b", "#030a12", "#171e26"
			case color1:
				return "#0e4429", "#001b00", "#002f12"
			case color2:
				return "#006d32", "#004307", "#00571b"
			case color3:
				return "#26a641", "#007d1a", "#11912e"
			case color4:
				return "#39d353", "#10a92c", "#24bd40"
			}
		}

		return "#2d333b", "#030a12", "#171e26"
	default:
		fallthrough
	case Light:
		if isHalloween {
			switch color {
			case halloweenColor0:
				return "#ebedf0", "#c2c5c8", "#d6d9dc"
			case halloweenColor1:
				return "#ffee4a", "#d5c522", "#e9d936"
			case halloweenColor2:
				return "#ffc501", "#d59d00", "#e9b100"
			case halloweenColor3:
				return "#fe9600", "#d56e00", "#e98200"
			case halloweenColor4:
				return "#03001c", "#000001", "#000007"
			}
		} else {
			switch color {
			case color0:
				return "#ebedf0", "#c2c5c8", "#d6d9dc"
			case color1:
				return "#9be9a8", "#73c080", "#87d494"
			case color2:
				return "#40c463", "#199b3c", "#2daf50"
			case color3:
				return "#30a14e", "#077725", "#1b8b39"
			case color4:
				return "#216e39", "#004410", "#0c5824"
			}
		}

		return "#ebedf0", "#c2c5c8", "#d6d9dc"
	}
}
