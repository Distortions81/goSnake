package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func main() {
	var startTiles = []XY{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}}
	players = append(players,
		playerData{ID: 1, Name: "Test", Color: 1, Length: 3, Tiles: startTiles, Head: 2, Tail: 0, Direction: DIR_EAST})

	startTiles = []XY{{X: 1, Y: 2}, {X: 1, Y: 2}, {X: 1, Y: 3}}
	players = append(players,
		playerData{ID: 1, Name: "Tester", Color: 2, Length: 3, Tiles: startTiles, Head: 2, Tail: 0, Direction: DIR_SOUTH})

	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowSize(int(boardSize*gridSize), int(boardSize*gridSize)+hudSize)

	go GameUpdate()

	if err := ebiten.RunGameWithOptions(newGame(), nil); err != nil {
		return
	}
}

func checkDir(dir uint8) bool {

	return false
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

		//fmt.Printf("tick %v\n", gameTick)

		gameLock.Unlock()
		sleepFor := sleepTime - int(time.Since(start).Nanoseconds())
		if sleepFor > 0 {
			time.Sleep(time.Duration(sleepFor))
		}
	}
}

func (g *Game) Update() error {
	gameLock.Lock()

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		players[0].Direction = DIR_NORTH
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		players[0].Direction = DIR_WEST
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		players[0].Direction = DIR_SOUTH
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		players[0].Direction = DIR_EAST
	}

	gameLock.Unlock()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	gameLock.Lock()

	for _, player := range players {
		for _, tile := range player.Tiles {
			if player.DeadFor > 0 {
				vector.DrawFilledRect(screen, float32((tile.X-1)*gridSize), float32((tile.Y-1)*gridSize), tileSize, tileSize, deadColor, false)
			} else {
				vector.DrawFilledRect(screen, float32((tile.X-1)*gridSize), float32((tile.Y-1)*gridSize), tileSize, tileSize, colorList[player.Color], false)
			}
		}
	}
	vector.DrawFilledRect(screen, 0, float32(screen.Bounds().Dy()-hudSize), float32(screen.Bounds().Dx()), hudSize, hudColor, false)
	buf := fmt.Sprintf("FPS: %0.2f, Players: %v", ebiten.ActualFPS(), len(players))
	ebitenutil.DebugPrintAt(screen, buf, 0, (screen.Bounds().Dy() - hudSize + 2))
	gameLock.Unlock()
}

func newGame() *Game {
	gameRunning = true
	return &Game{}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
