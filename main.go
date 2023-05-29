package main

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	players     []playerData
	gameTick    uint64
	gameLock    sync.Mutex
	gameRunning bool
)

type XY struct {
	X int16
	Y int16
}

const (
	startGameDelayMS = 1000
	gridSize         = 16
	border           = 1
	tileSize         = gridSize - border
	boardSize        = 50
	gameSpeed        = 4

	DIR_NONE  = 0
	DIR_NORTH = 1
	DIR_EAST  = 2
	DIR_SOUTH = 3
	DIR_WEST  = 4
)

type playerData struct {
	Name   string
	Color  uint8
	Length uint32
	ID     uint32

	Tiles []XY
	Head  uint32
	Tail  uint32

	Direction uint8
	DeadFor   uint8

	Command uint8
}

type Game struct {
}

func main() {
	var startTiles = []XY{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}}
	players = append(players,
		playerData{ID: 1, Name: "Test", Color: 1, Length: 3, Tiles: startTiles, Head: 2, Tail: 0, Direction: DIR_EAST})

	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowSize(int(boardSize*gridSize), int(boardSize*gridSize))

	go GameUpdate()

	if err := ebiten.RunGameWithOptions(newGame(), nil); err != nil {
		return
	}
}

func goDir(dir uint8, pos XY) XY {
	switch dir {
	case DIR_NORTH:
		pos.Y--
	case DIR_EAST:
		pos.X++
	case DIR_SOUTH:
		pos.Y++
	case DIR_WEST:
		pos.X--
	}
	return pos
}

func GameUpdate() {
	sleepTime := 1000000000 / gameSpeed
	gameTick = 0

	for !gameRunning {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Millisecond * startGameDelayMS)

	for gameRunning {
		start := time.Now()
		gameTick++
		gameLock.Lock()

		deletePlayer := -1
		for p, player := range players {
			if player.DeadFor > 0 {
				players[p].DeadFor++
				if player.DeadFor > 4 {
					deletePlayer = p
				}
				continue
			}
			head := player.Tiles[player.Head]
			newHead := goDir(player.Direction, head)
			if newHead.X > boardSize || newHead.Y > boardSize ||
				newHead.X < 1 || newHead.Y < 1 {
				players[p].DeadFor = 1
				fmt.Printf("Player %v #%v died.\n", player.Name, player.ID)
				continue
			}

			players[p].Tiles = append(player.Tiles[1:], XY{X: newHead.X, Y: newHead.Y})
			players[p].Head = player.Length - 1

		}
		if deletePlayer > -1 {
			fmt.Printf("Player %v #%v deleted.\n", players[deletePlayer].Name, players[deletePlayer].ID)
			players = append(players[:deletePlayer], players[deletePlayer+1:]...)
		}

		fmt.Printf("tick %v\n", gameTick)

		gameLock.Unlock()
		sleepFor := sleepTime - int(time.Since(start).Nanoseconds())
		if sleepFor > 0 {
			time.Sleep(time.Duration(sleepFor))
		}
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	gameLock.Lock()

	for _, player := range players {
		for _, tile := range player.Tiles {
			if player.DeadFor > 0 {
				vector.DrawFilledRect(screen, float32((tile.X-1)*gridSize), float32((tile.Y-1)*gridSize), tileSize, tileSize, color.NRGBA{0xFF, 0, 0, 0xFF}, false)
			} else {
				vector.DrawFilledRect(screen, float32((tile.X-1)*gridSize), float32((tile.Y-1)*gridSize), tileSize, tileSize, colorList[player.Color], false)
			}
		}
	}

	gameLock.Unlock()
}

func newGame() *Game {
	gameRunning = true
	return &Game{}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

var colorList = []color.NRGBA{
	{255, 255, 255, 255},
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
