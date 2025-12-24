package geo

import "math"

func toCoordinatePair(value interface{}) (float64, float64, bool) {
	pair, ok := value.([]interface{})
	if !ok || len(pair) < 2 {
		return 0, 0, false
	}
	lon, okLon := pair[0].(float64)
	lat, okLat := pair[1].(float64)
	if !okLon || !okLat {
		return 0, 0, false
	}
	return lon, lat, true
}

func parseRing(value interface{}) ([][2]float64, bool) {
	ringSlice, ok := value.([]interface{})
	if !ok {
		return nil, false
	}
	var ring [][2]float64
	for _, coord := range ringSlice {
		lon, lat, ok := toCoordinatePair(coord)
		if !ok {
			continue
		}
		ring = append(ring, [2]float64{lon, lat})
	}
	if len(ring) == 0 {
		return nil, false
	}
	return ring, true
}

func extractCoordinates(geometry map[string]interface{}) [][][2]float64 {
	// ジオメトリがnilの場合はnilを返す
	if geometry == nil {
		return nil
	}

	// ジオメトリタイプと座標の取得
	geomTypeField := geometry["type"]
	geomType, _ := geomTypeField.(string)
	coordsField, ok := geometry["coordinates"]
	if !ok {
		return nil
	}
	coordsSlice, ok := coordsField.([]interface{})
	if !ok || len(coordsSlice) == 0 {
		return nil
	}

	// ジオメトリタイプに応じた座標の抽出
	switch geomType {
	case "Point":
		if lon, lat, ok := toCoordinatePair(coordsSlice); ok {
			return [][][2]float64{{{lon, lat}}}
		}
	case "Polygon":
		if ring, ok := parseRing(coordsSlice[0]); ok {
			return [][][2]float64{ring}
		}
	case "MultiPolygon":
		var result [][][2]float64
		for _, poly := range coordsSlice {
			polyRings, ok := poly.([]interface{})
			if !ok || len(polyRings) == 0 {
				continue
			}
			if ring, ok := parseRing(polyRings[0]); ok {
				result = append(result, ring)
			}
		}
		return result
	}

	return nil
}

func calculateBoundingBox(features []map[string]interface{}) *Bound {
	positiveInf := math.Inf(1)
	negativeInf := math.Inf(-1)

	bound := &Bound{
		LonMin: positiveInf,
		LonMax: negativeInf,
		LatMin: positiveInf,
		LatMax: negativeInf,
	}

	// ジオメトリの抽出とバウンディングボックスの更新
	for _, feature := range features {
		geometryField, ok := feature["geometry"]
		if !ok {
			continue
		}
		geometry, ok := geometryField.(map[string]interface{})
		if !ok {
			continue
		}
		coordinates := extractCoordinates(geometry)
		for _, rings := range coordinates {
			for _, ring := range rings {
				lon, lat := ring[0], ring[1]
				bound.LonMin = math.Min(bound.LonMin, lon)
				bound.LonMax = math.Max(bound.LonMax, lon)
				bound.LatMin = math.Min(bound.LatMin, lat)
				bound.LatMax = math.Max(bound.LatMax, lat)
			}
		}
	}
	return bound

}
