/*
# convert.go

地理座標（経度緯度）をターミナルUI座標（X, Y）に変換するモジュール
*/
package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
)

func BytesToLayer(data []byte) (Layer, error) {
	// データが空の場合はエラーを返す
	if len(data) == 0 {
		return Layer{
			Valid: false,
		}, errors.New("empty GeoJSON Bytes")
	}

	// jsonのパース
	var geojson map[string]interface{}
	if err := json.Unmarshal(data, &geojson); err != nil {
		return Layer{
			Valid: false,
		}, fmt.Errorf("parse JSON: %w", err)
	}

	// Layer型に変換
	layer, err := ConvertTuiLayer(geojson)
	if err != nil {
		return Layer{
			Valid: false,
		}, fmt.Errorf("convert to Layer: %w", err)
	}

	return layer, nil
}

/*
地理座標（経度緯度）をターミナルUI座標（X, Y）に変換する。
【変換ロジック】
X軸（東西）:
- GeoJSONの経度: [lon_min, lon_max]
- TUI座標: [0, width-1]
- 線形補間で変換
Y軸（南北）:
- GeoJSONの緯度: [lat_min, lat_max]（北が大きい値）
- TUI座標: [0, height-1]（上がY=0）
- 上下反転が必要: y = height - 1 - normalized_y

Args:

	lon: 経度（例: 135.5）
	lat: 緯度（例: 34.5）
	bounds: 全体のBounding Box
	width: ターミナル幅（セル数）
	height: ターミナル高さ（セル数）

Returns:

	[x, y] のスライス。範囲は [0, width-1] × [0, height-1]
*/
func geometoryToTui(lon, lat float64, bound *Bound, width, height int) [2]int {
	// 経度の正規化
	var xNorm, yNorm float64
	if bound.lonSpan() > 0 {
		xNorm = (lon - bound.LonMin) / bound.lonSpan() * float64(width-1)
	} else {
		xNorm = float64(width-1) / 2
	}
	x := int(math.Round(xNorm))

	// 緯度の正規化（上下反転）
	if bound.latSpan() > 0 {
		yNorm = (bound.LatMax - lat) / bound.latSpan() * float64(height-1)
	} else {
		yNorm = float64(height-1) / 2
	}
	y := int(math.Round(yNorm))

	if x < 0 {
		x = 0
	} else if x > width-1 {
		x = width - 1
	}

	if y < 0 {
		y = 0
	} else if y > height-1 {
		y = height - 1
	}

	return [2]int{x, y}
}

/*
ConvertTui
GeoJSONファイルを読み込み、地理座標をターミナルUI座標に変換する。
Args:

	path: GeoJSONファイルのパス
	width: ターミナル幅（セル数）
	height: ターミナル高さ（セル数）

Returns:

	TuiGeometry
*/

func ConvertTui(path string, width, height int) (TuiGeometry, error) {
	// geojsonファイルの読み込み
	data, err := os.ReadFile(path)
	if err != nil {
		return TuiGeometry{}, fmt.Errorf("read file: %w", err)
	}

	// jsonのパース
	layer, err := BytesToLayer(data)
	if err != nil {
		return TuiGeometry{}, fmt.Errorf("parse JSON: %w", err)
	}

	return ConvertTuiBytes(layer, width, height)
}

/*
ConvertTuiBytes
パース済みのgeojsonデータを読み込み、地理座標をターミナルUI座標に変換する。

Args:

	data: パース済みのGeoJSONデータ
	width: ターミナル幅（セル数）
	height: ターミナル高さ（セル数）

Returns:

	TuiGeometry
*/
func ConvertTuiBytes(layer Layer, width, height int) (TuiGeometry, error) {
	// 各featureの処理
	var polygons []Polygon
	for _, feature := range layer.Features {
		// 各ringsの処理
		rings := feature.Rings
		var tuiRings [][][2]int
		for _, ring := range rings {
			var tuiRing [][2]int
			for _, coord := range ring {
				lon, lat := coord[0], coord[1]
				tuiCoord := geometoryToTui(lon, lat, &layer.Bounds, width, height)
				tuiRing = append(tuiRing, tuiCoord)
			}
			tuiRings = append(tuiRings, tuiRing)
		}

		polygon := Polygon{
			Name:       feature.Name,
			Properties: feature.Properties,
			Rings:      tuiRings,
		}
		polygons = append(polygons, polygon)
	}

	return TuiGeometry{
		Bounds:   layer.Bounds,
		Width:    width,
		Height:   height,
		Polygons: polygons,
	}, nil
}

// ConvertTuiLayer
// パース済みのgeojsonデータを読み込み、地理座標をLayer型で返す。
//
// Args:
//
//	data: パース済みのGeoJSONデータ
//
// Returns:
//
//	Layer
func ConvertTuiLayer(geojson map[string]interface{}) (Layer, error) {
	// geojsonがnilの場合はエラーを返す
	if geojson == nil {
		return Layer{
			Valid: false,
		}, errors.New("geojson is nil")
	}

	// featuresの取得
	featuresInterface, ok := geojson["features"]
	if !ok {
		return Layer{
			Valid: false,
		}, errors.New("features not found in GeoJSON")
	}
	// 型アサーション
	features, ok := featuresInterface.([]interface{})
	if !ok {
		return Layer{
			Valid: false,
		}, errors.New("features is not a slice")
	}
	// featuresをmap[string]interface{}のスライスに変換
	var featureMaps []map[string]interface{}
	for _, feature := range features {
		featureMap, ok := feature.(map[string]interface{})
		if !ok {
			continue
		}
		featureMaps = append(featureMaps, featureMap)
	}

	// バウンディングボックスの計算
	bound := calculateBoundingBox(featureMaps)
	if math.IsInf(bound.LonMin, 0) || math.IsInf(bound.LatMin, 0) {
		return Layer{
			Valid: false,
		}, errors.New("bounding box could not be calculated")
	}

	// 各featureの処理
	var layerFeatures []CachedFeature
	for _, feature := range featureMaps {
		// geometryの取得
		// 型アサーションしてmap[string]interface{}に変換
		geometryField, ok := feature["geometry"]
		if !ok {
			continue
		}
		geometry, ok := geometryField.(map[string]interface{})
		if !ok {
			continue
		}

		// propertiesの取得
		propertiesField, ok := feature["properties"]
		if !ok {
			continue
		}
		properties, ok := propertiesField.(map[string]interface{})
		if !ok {
			continue
		}

		rings := extractCoordinates(geometry)

		name, ok := properties["name"]
		if !ok {
			name = "unknown"
		}

		cachedFeature := CachedFeature{
			Name:       name.(string),
			Properties: properties,
			Rings:      rings,
		}
		layerFeatures = append(layerFeatures, cachedFeature)
	}

	return Layer{
		Bounds:   *bound,
		Features: layerFeatures,
		Valid:    true,
	}, nil
}
