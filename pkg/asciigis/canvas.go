package asciigis

import (
	"asciigis/internal/geo"
	"fmt"
)

func canvas(geometry geo.TuiGeometry) {
	// 空のキャンバス作成
	var canvas [][]rune
	for y := 0; y < geometry.Height; y++ {
		row := make([]rune, geometry.Width)
		for x := 0; x < geometry.Width; x++ {
			row[x] = ' ' // 空白で初期化
		}
		canvas = append(canvas, row)
	}

	// 各ポリゴンの描画
	for _, polygon := range geometry.Polygons {
		for _, ring := range polygon.Rings {
			for _, coord := range ring {
				x, y := coord[0], coord[1]
				canvas[y][x] = '*' // ポリゴンの点を'*'で描画
			}
		}
	}

	// キャンバスの表示
	for _, row := range canvas {
		fmt.Println(string(row))
	}
}
