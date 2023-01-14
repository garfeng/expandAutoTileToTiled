package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

var (
	//srcRoot = flag.String("input", "input", "src root contains all images")
	//dstRoot = flag.String("output", "output", "dst root to save generated tilemaps")
	isDebug = flag.Bool("debug", false, "is debug")
)

//go:embed all:maps
var embedMap embed.FS

const embedMapDir = "maps"

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("usage:\r\nexpandAutoTileToTiled.exe <inputTilesetsDirectory> [-debug=false]")
		pause()
		return
	}

	engine := &Engine{
		IsDebug: *isDebug,
		SrcRoot: os.Args[1],
		DstRoot: "./expandOutput",
	}
	engine.Generate()

	pause()
}

func pause() {
	b := ""
	fmt.Println("Finished, please close the program")
	fmt.Scanf("%s", &b)
}

type Engine struct {
	IsDebug bool
	SrcRoot string
	DstRoot string
}

const (
	baseDir = "./tmpMaps"
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
			fmt.Println("Fail to load map:", err)
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

	tileNumberWidthMap := map[string]int{
		"1": 16,
		"2": 16,
		"3": 16,
		"4": 16,
	}

	for _, v := range names {
		lowerV := strings.ToLower(v)
		_, name := filepath.Split(v)
		if filepath.Ext(lowerV) == ".png" {
			matched := autoNameRegexp.FindAllStringSubmatch(lowerV, -1)
			if len(matched) > 0 {
				srcPath := filepath.Join(e.SrcRoot, v)
				typeId := matched[0][1]
				img, err := readImage(srcPath)
				if err != nil {
					fmt.Println("Fail to read image", srcPath, "Error:", err)
					continue
				}
				tn := tileNumberWidthMap[typeId]
				tileSize := float64(img.Bounds().Dx()) / float64(tn)
				res = append(res, TileConfig{
					SrcImagePath:      srcPath,
					TempLoadImagePath: fmt.Sprintf("%s/A%s.png", baseDir, typeId),

					Tileset:         fmt.Sprintf("%s/A%s.tsx", embedMapDir, typeId),
					TempLoadTileset: fmt.Sprintf("%s/A%s.tsx", baseDir, typeId),

					AutoMapName:    fmt.Sprintf("%s/A%s_expand.tmx", embedMapDir, typeId),
					Name:           name,
					TileSize:       int(tileSize),
					SrcImageWidth:  img.Bounds().Dx(),
					SrcImageHeight: img.Bounds().Dy(),
				})
			}
		}
	}
	return res, nil
}

func readImage(name string) (image.Image, error) {
	buff, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(buff)
	return png.Decode(r)
}

func (e *Engine) PrepareTileData(cfg *TileConfig) error {
	err := copyFile(cfg.SrcImagePath, cfg.TempLoadImagePath)
	if err != nil {
		return err
	}

	if e.IsDebug {
		err = copyFile(cfg.Tileset, cfg.TempLoadTileset)
	} else {
		var buff []byte
		buff, err = embedMap.ReadFile(cfg.Tileset)
		if err == nil {
			err = os.WriteFile(cfg.TempLoadTileset, buff, 0755)
		}
	}

	//err = copyFile(cfg.Tileset, cfg.TempLoadTileset)
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
	gameMap.TileWidth = cfg.TileSize / 2
	gameMap.TileHeight = cfg.TileSize / 2

	for _, tileset := range gameMap.Tilesets {
		tileset.TileWidth = cfg.TileSize / 2
		tileset.TileHeight = cfg.TileSize / 2
		tileset.Image.Width = cfg.SrcImageWidth
		tileset.Image.Height = cfg.SrcImageHeight
	}

	fmt.Println("tileSize = ", cfg.TileSize)

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

	tileset, err := loadJSON[Tileset](dstTileBuff)
	if err != nil {
		return err
	}

	(*dstMap)["image"] = filepath.Join("../../img/tilesets", dstImageName)
	(*dstMap)["name"] = replaceExtTo(dstImageName, "")
	(*dstMap)["imagewidth"] = int(tileset.Columns) * cfg.TileSize
	(*dstMap)["imageheight"] = int(tileset.Imageheight/tileset.Tileheight) * cfg.TileSize
	(*dstMap)["tilewidth"] = cfg.TileSize
	(*dstMap)["tileheight"] = cfg.TileSize

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

	Tileset         string
	TempLoadTileset string

	AutoMapName string
	Name        string

	TileSize int

	SrcImageWidth  int
	SrcImageHeight int
}
