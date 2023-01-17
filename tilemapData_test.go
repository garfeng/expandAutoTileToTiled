package main

import (
	"encoding/json"
	"os"
	"testing"
)

func Test_ParseTilesetA1(t *testing.T) {
	buff, _ := os.ReadFile("./temp/A1.json")
	tileset, _ := loadJSON[Tileset](buff)

	for i, v := range tileset.Tiles {
		tileset.Tiles[i].Animation = []Animation{}
		for j := 0; j < 3; j++ {
			tileset.Tiles[i].Animation = append(tileset.Tiles[i].Animation, Animation{
				Duration: 100,
				Tileid:   v.ID + int64(j)*7,
			})
		}
	}

	buff2, _ := json.MarshalIndent(tileset, "", "  ")
	os.WriteFile("./temp/A1_dst.json", buff2, 0755)
}

func Test_ParseTilesetA3(t *testing.T) {
	buff, _ := os.ReadFile("./temp/A3.json")
	tileset, _ := loadJSON[Tileset](buff)
	wangsets := []Wangset{}

	for j := 0; j < 4; j += 2 {
		for i := 0; i < 8; i++ {
			idx := j*8 + i
			w1 := tileset.Wangsets[idx]
			w2 := tileset.Wangsets[idx+8]
			wangsets = append(wangsets, w1, w2)
		}
	}
	tileset.Wangsets = wangsets

	buff2, _ := json.MarshalIndent(tileset, "", "  ")
	os.WriteFile("./temp/A3_dst.json", buff2, 0755)
}
