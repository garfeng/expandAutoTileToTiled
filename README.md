# Expand RPG Maker AutoTile to Tiled style

The program is aimed for [Tiled Plugin for RPG Maker MZ by VisuStella, VisuStellaMZ, Archeia (itch.io)](https://visustella.itch.io/tiledpluginmz)

Download: https://github.com/garfeng/expandAutoTileToTiled/releases

| Item               | Images                                                       | Info |
| ------------------ | ------------------------------------------------------------ | ---- |
| Inputs<br>(RMMZ)   | ![image-20230114230711059](README.assets/image-20230114230711059.png) |      |
| Outputs<br>(Tiled) | ![image-20230114230757551](README.assets/image-20230114230757551.png) |      |
| Terrains           | ![image-20230114230909483](README.assets/image-20230114230909483.png) |      |
| Animations         | ![image-20230114231054718](README.assets/image-20230114231054718.png) |      |



## Functions

- [x] Expand images
- [x] Import terrains
- [x] Import animations

- [x] Tile size of 48/32/24/16 supported.



## Usage

Drag The directory with AutoTile tilesets in it to `expandAutoTileToTiled.exe`, the output images and tilesets will be created in `expandOutput`

For example, drag the `tilesets` directory in `YourProject/img` , to `expandAutoTileToTiled.exe`

![image-20230114231608690](README.assets/image-20230114231608690.png)

The program detects auto tile with **Image Name**, that means `A1~A4` is required, please put `A1~A4` in your auto tile image names.



## Licenses

Free for commercial and noncommercial games.



## Thanks to

* [eishiya/tiled-scripts: Assorted scripts for Tiled Map Editor. (github.com)](https://github.com/eishiya/tiled-scripts)

* [lafriks/go-tiled: Go library to parse Tiled map editor file format (TMX) and render map to image (github.com)](https://github.com/lafriks/go-tiled)

* [Tiled | Flexible level editor (mapeditor.org)](https://www.mapeditor.org/)

* [Tiled Plugin for RPG Maker MZ by VisuStella, VisuStellaMZ, Archeia (itch.io)](https://visustella.itch.io/tiledpluginmz)
