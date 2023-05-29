package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var players []playerData

type XY struct {
	X int16
	Y int16
}

type playerData struct {
	Name     string
	Color    uint8
	Length   uint16
	Tiles    []XY
	LastTile uint16

	Command uint8
}

type Game struct {
}

func main() {
	var startTiles = []XY{{X: 1, Y: 1}}
	players = append(players, playerData{Name: "test", Color: 1, Length: 1, Tiles: startTiles})

	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowSize(int(boardSize*gridSize), int(boardSize*gridSize))

	if err := ebiten.RunGameWithOptions(newGame(), nil); err != nil {
		return
	}

	time.Sleep(time.Second)
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, player := range players {
		for _, tile := range player.Tiles {
			vector.DrawFilledRect(screen, float32(tile.X), float32(tile.Y), tileSize, tileSize, colorList[player.Color], false)
		}
	}

}

func newGame() *Game {
	return &Game{}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

const gridSize = 32
const border = 1
const tileSize = gridSize - border
const boardSize = 20

var colorList = []color.NRGBA{
	{203, 67, 53, 255},
	{40, 180, 99, 255},
	{41, 128, 185, 255},
	{244, 208, 63, 255},
	{243, 156, 18, 255},
	{255, 151, 197, 255},
	{165, 105, 189, 255},
	{209, 209, 209, 255},
	{64, 199, 178, 255},
	{199, 54, 103, 255},
	{99, 114, 166, 255},
	{134, 166, 99, 255},
	{206, 231, 114, 255},
	{209, 114, 231, 255},
	{114, 228, 231, 255},
	{176, 116, 78, 255},
	{210, 113, 52, 255},
}
