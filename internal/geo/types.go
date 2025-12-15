package geo

type Polygon struct {
	Name       string
	Properties map[string]interface{}
	Rings      [][][2]int // TUI座標系でのリング
}

type Bound struct {
	LonMin, LonMax float64
	LatMin, LatMax float64
}

func (b *Bound) latSpan() float64 {
	return b.LatMax - b.LatMin
}

func (b *Bound) lonSpan() float64 {
	return b.LonMax - b.LonMin
}

type TuiGeometry struct {
	Bounds   Bound     `json:"bounds"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Polygons []Polygon `json:"polygons"`
}
