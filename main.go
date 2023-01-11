package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

var (
	srcRoot = flag.String("src", "src", "src root contains all images")
	dstRoot = flag.String("dst", "dst", "dst root to save generated tilemaps")
	isDebug = flag.Bool("debug", false, "is debug")
)

//go:embed all:maps
var embedMap embed.FS

const embedMapDir = "maps"

func main() {
	flag.Parse()

	engine := &Engine{
		IsDebug: *isDebug,
		SrcRoot: *srcRoot,
		DstRoot: *dstRoot,
	}
	engine.Generate()
}

type Engine struct {
	IsDebug bool
	SrcRoot string
	DstRoot string
}

const (
	baseDir = "./maps"
)

func init() {
	os.MkdirAll(baseDir, 0755)
}

func (e *Engine) Generate() error {
	err := os.MkdirAll(e.TilesetImageDstRoot(), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(e.TilesetJSONDstRoot(), 0755)
	if err != nil {
		return err
	}

	data, err := e.ScanAutoTiles()
	if err != nil {
		return err
	}

	for _, v := range data {
		fmt.Println("Parse", v.Name)
		err := e.PrepareTileData(&v)
		if err != nil {
			fmt.Println(err)
			continue
		}
		gameMap, err := e.LoadMap(v.AutoMapName)
		if err != nil {
			fmt.Println(err)
			continue
		}

		e.GenerateOneMap(gameMap, &v)
	}
	return nil
}

func (e *Engine) LoadMap(mapPath string) (*tiled.Map, error) {
	buff, err := e.LoadMapBuff(mapPath)
	if err != nil {
		return nil, err
	}

	_, name := filepath.Split(mapPath)

	newMapPath := filepath.Join(baseDir, name)
	err = os.WriteFile(newMapPath, buff, 0755)
	if err != nil {
		return nil, err
	}

	return tiled.LoadFile(newMapPath)
}

func (e *Engine) LoadMapBuff(mapPath string) ([]byte, error) {
	var buff []byte
	var err error

	if e.IsDebug {
		buff, err = os.ReadFile(mapPath)
	} else {
		buff, err = embedMap.ReadFile(mapPath)
	}

	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (e *Engine) ScanAutoTiles() ([]TileConfig, error) {
	fp, err := os.Open(e.SrcRoot)
	if err != nil {
		return nil, err
	}
	names, err := fp.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	res := []TileConfig{}

	for _, v := range names {
		lowerV := strings.ToLower(v)
		_, name := filepath.Split(v)
		if filepath.Ext(lowerV) == ".png" {
			matched := autoNameRegexp.FindAllStringSubmatch(lowerV, -1)
			if len(matched) > 0 {
				srcPath := filepath.Join(e.SrcRoot, v)
				typeId := matched[0][1]
				res = append(res, TileConfig{
					SrcImagePath:        srcPath,
					TempLoadImagePath:   fmt.Sprintf("%s/A%s.png", baseDir, typeId),
					SrcTilemapName:      fmt.Sprintf("%s/A%s.tsx", embedMapDir, typeId),
					TempLoadTilemapName: fmt.Sprintf("%s/A%s.tsx", baseDir, typeId),
					AutoMapName:         fmt.Sprintf("%s/A%s_expand.tmx", embedMapDir, typeId),

					Name: name,
				})
			}
		}
	}
	return res, nil
}

func (e *Engine) PrepareTileData(cfg *TileConfig) error {
	err := copyFile(cfg.SrcImagePath, cfg.TempLoadImagePath)
	if err != nil {
		return err
	}

	err = copyFile(cfg.SrcTilemapName, cfg.TempLoadTilemapName)
	if err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	buff, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, buff, 0755)
}

func (e *Engine) TilesetImageDstRoot() string {
	return filepath.Join(e.DstRoot, "img", "tilesets")
}

func (e *Engine) TilesetJSONDstRoot() string {
	return filepath.Join(e.DstRoot, "maps", "Tilesets")
}

func (e *Engine) GenerateOneMap(gameMap *tiled.Map, cfg *TileConfig) error {
	for i := range gameMap.Layers {
		err := e.GenerateOneLayer(gameMap, i, cfg)
		if err != nil {
			fmt.Printf("Fail to generate %s, layer %d, err : %s\r\n", cfg.SrcImagePath, i+1, err.Error())
		}
	}
	return nil
}

func (e *Engine) GenerateOneLayer(gameMap *tiled.Map, layerIdx int, cfg *TileConfig) error {
	renderer, err := render.NewRenderer(gameMap)
	if err != nil {
		return err
	}

	// Render just layer 0 to the Renderer.
	err = renderer.RenderLayer(layerIdx)
	if err != nil {
		return err
	}
	defer renderer.Clear()

	// Get a reference to the Renderer's output, an image.NRGBA struct.
	img := renderer.Result
	w := bytes.NewBuffer(nil)
	err = png.Encode(w, img)
	if err != nil {
		return err
	}

	dstImageName := replaceExtTo(cfg.Name, fmt.Sprintf("_%d.png", layerIdx+1))
	dstImagePath := filepath.Join(e.TilesetImageDstRoot(), dstImageName)

	err = os.WriteFile(dstImagePath, w.Bytes(), 0755)
	if err != nil {
		return err
	}

	layer := gameMap.Layers[layerIdx]

	tilemapName := layer.Properties.GetString("tileMapName")
	if tilemapName == "" {
		return errors.New("can not find property name \"tileMapName\"")
	}

	dstTileBuff, err := e.LoadMapBuff(tilemapName)
	dstMap, err := loadJSON[TileMapData](dstTileBuff)
	if err != nil {
		return err
	}

	(*dstMap)["image"] = filepath.Join("../../img/tilesets", dstImageName)
	(*dstMap)["name"] = replaceExtTo(dstImageName, "")

	dstTileBuff, err = json.MarshalIndent(dstMap, "", "  ")
	if err != nil {
		return err
	}
	dstTileName := replaceExtTo(cfg.Name, fmt.Sprintf("_%d.json", layerIdx+1))
	dstTilePath := filepath.Join(e.TilesetJSONDstRoot(), dstTileName)
	err = os.WriteFile(dstTilePath, dstTileBuff, 0755)
	return err
}

func replaceExtTo(name string, dstExt string) string {
	ext := filepath.Ext(name)
	idx := strings.LastIndex(name, ext)
	return name[:idx] + dstExt
}

func loadJSON[T any](buff []byte) (*T, error) {
	v := new(T)
	err := json.Unmarshal(buff, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

var (
	autoNameRegexp = regexp.MustCompile(`a([1-4])`)
)

type MapConfig struct {
}

type TileConfig struct {
	SrcImagePath      string
	TempLoadImagePath string

	SrcTilemapName      string
	TempLoadTilemapName string

	AutoMapName string
	Name        string
}
