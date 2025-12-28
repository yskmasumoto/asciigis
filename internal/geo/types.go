package geo

// geojsonを最初に読んだ後、内部で保持する際の型定義
// width, heightが変化したとき、この型からTuiGeometryに変換する
type Layer struct {
	// 地理座標系での境界ボックス
	Bounds Bound
	// 各フィーチャー
	Features []CachedFeature
	// パースが有効かどうか
	Valid bool `json:"valid"`
}

type CachedFeature struct {
	Name       string
	Properties map[string]interface{}
	Rings      [][][2]float64 // 経度緯度座標系でのリング
}

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
