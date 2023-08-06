package graph

// Theme is the theme that is used to render the graph (Light or Dark).
type Theme int

const (
	Dark Theme = iota
	Light
)

type vector2 struct {
	X float64
	Y float64
}

type vector3 struct {
	X float64
	Y float64
	Z float64
}

type col struct {
	pts   []vector3
	count int
	color string
}
