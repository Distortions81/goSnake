package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	TouchMode    bool = false
	TouchPressed bool
	PinchPressed bool
	ZoomMouse    float64
	LastPinch    float64
	CameraX      float64
	CameraY      float64
	ZoomScale    float64 = 1.0

	ScrollPos  float64
	LastTouchA int
	LastTouchB int
	LastTouchX int
	LastTouchY int
	lastScroll time.Time

	GoDirThisTick int

	arrowIconSize XY16
	arrowIconPos  XY16
	arrowScale    float32
)

func updateTouchControls() {
	if verticalMode {
		/* Icon starts after game board */
		startY := float64(boardPixels + (gridSize))

		/* Calculate scale needed to fit image */
		scaleY := 0.9 / (float32(touchIcon.Bounds().Dy()) / (float32(ScreenHeight) - float32(startY)))
		scaleX := 0.9 / (float32(touchIcon.Bounds().Dx()) / (float32(ScreenWidth)))

		/* Keep icon from going off-screen in extreme aspect ratios */
		if scaleX < scaleY {
			arrowScale = scaleX
		} else {
			arrowScale = scaleY
		}

		/* Calculate width at this scale */
		width := float32(touchIcon.Bounds().Dx()) * arrowScale
		height := float32(touchIcon.Bounds().Dy()) * arrowScale

		/* Start sprite at end of game board, then center it horizontally */
		spriteX := (float64(ScreenWidth) - float64(width)) / 2.0
		spriteY := ScreenHeight - int(height)

		/* Save these values for ui.go */
		arrowIconPos.X = uint16(spriteX)
		arrowIconPos.Y = uint16(spriteY)
		arrowIconSize.X = uint16(width)
		arrowIconSize.Y = uint16(height)
	} else {
		/* Icon starts after game board */
		startX := float64(boardPixels + (gridSize * 2))

		/* Calculate scale needed to fit image */
		scaleY := 0.8 / (float32(touchIcon.Bounds().Dy()) / (float32(ScreenHeight)))
		scaleX := 0.8 / (float32(touchIcon.Bounds().Dx()) / (float32(ScreenWidth) - float32(startX)))

		/* Keep icon from going off-screen in extreme aspect ratios */
		if scaleX < scaleY {
			arrowScale = scaleX
		} else {
			arrowScale = scaleY
		}

		/* Calculate width at this scale */
		width := float32(touchIcon.Bounds().Dx()) * arrowScale
		height := float32(touchIcon.Bounds().Dy()) * arrowScale

		/* Start sprite at end of game board, then center it horizontally */
		spriteX := ScreenWidth - int(width)
		spriteY := ScreenHeight - int(height)

		/* Save these values for ui.go */
		arrowIconPos.X = uint16(spriteX)
		arrowIconPos.Y = uint16(spriteY)
		arrowIconSize.X = uint16(width)
		arrowIconSize.Y = uint16(height)
	}
}

/* Input interface handler */
func (g *Game) Update() error {

	//Lock/Unlock
	gameLock.Lock()
	defer gameLock.Unlock()

	if audioPlayer != nil && !audioPlayer.IsPlaying() {
		audioPlayer.Rewind()
		audioPlayer.SetVolume(0.20)
		audioPlayer.Play()
		doLog(true, "Playing music...")
	}

	/* Ignore if game not focused */
	if !ebiten.IsFocused() {
		return nil
	}

	/* Clamp to window */
	MouseX, MouseY = ebiten.CursorPosition()
	if MouseX < 0 || MouseX > int(ScreenWidth) ||
		MouseY < 0 || MouseY > int(ScreenHeight) {
		MouseX = lastMouseX
		MouseY = lastMouseY
	}
	// Touchscreen input
	tids := inpututil.AppendJustPressedTouchIDs(nil)

	tx := 0
	ty := 0
	ta := 0
	tb := 0

	/* Find touch events */
	foundTouch := false
	foundPinch := false
	for _, tid := range tids {
		ttx, tty := ebiten.TouchPosition(tid)
		if ttx > 0 || tty > 0 {
			TouchMode = true
			if foundTouch {
				ta = ttx
				tb = tty
				foundPinch = true
				break
			} else {
				tx = ttx
				ty = tty

				//Move mouse
				MouseX = tx
				MouseY = ty
				foundTouch = true
				break
			}

		}
	}

	/* Touch zoom-pinch */
	if foundPinch {
		dist := distance((ta), (tb), (tx), (ty))
		if !PinchPressed {
			LastPinch = dist
		}
		PinchPressed = true
		ZoomMouse = (ZoomMouse + ((dist - LastPinch) / 75))
		LastPinch = dist
		TouchPressed = false
	} else {
		if PinchPressed {
			TouchPressed = false
			foundTouch = false
		}
		PinchPressed = false
	}
	/* Touch pan */
	if foundTouch {
		if !TouchPressed {
			if PinchPressed {
				LastTouchA, LastTouchB = midPoint(tx, ty, ta, tb)

			} else {
				LastTouchX = tx
				LastTouchY = ty
			}
		}
		TouchPressed = true

		if PinchPressed {
			nx, ny := midPoint(tx, ty, ta, tb)
			CameraX = CameraX + (float64(LastTouchA-nx) / ZoomScale)
			CameraY = CameraY + (float64(LastTouchB-ny) / ZoomScale)
			LastTouchA, LastTouchB = midPoint(tx, ty, ta, tb)
		} else {
			CameraX = CameraX + (float64(LastTouchX-tx) / ZoomScale)
			CameraY = CameraY + (float64(LastTouchY-ty) / ZoomScale)
			LastTouchX = tx
			LastTouchY = ty
		}
	} else {
		TouchPressed = false
		CameraY = 0
	}

	if CameraY > float64(itemHeight) {
		if scrollOffset < len(lobbies)-1 {
			scrollOffset++
			CameraY = 0
		}
	} else if CameraY < -float64(itemHeight) {
		if scrollOffset > 0 {
			scrollOffset--
			CameraY = 0
		}
	}

	/*************
	 * GAME MODE *
	 *************/
	if gameMode == MODE_PLAY_GAME {

		var dir DIR = 0xFF

		if TouchMode && (TouchPressed || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)) && localPlayer.deadFor == 0 {
			/* Up Arrow */
			if MouseX > int(arrowIconPos.X+(arrowIconSize.X/3)) &&
				MouseX < int(arrowIconPos.X+((arrowIconSize.X/3)*2)) &&
				MouseY > int(arrowIconPos.Y) &&
				MouseY < int(arrowIconPos.Y+(arrowIconSize.Y/2)) {
				dir = DIR_NORTH
				keyPressed = dir
				//lastKeyPressed = time.Now()

				/* Left Arrow */
			} else if MouseX > int(arrowIconPos.X) &&
				MouseX < int(arrowIconPos.X+(arrowIconSize.X/3)) &&
				MouseY > int(arrowIconPos.Y+(arrowIconSize.Y/2)) &&
				MouseY < int(arrowIconPos.Y+(arrowIconSize.Y)) {
				dir = DIR_WEST
				keyPressed = dir
				//lastKeyPressed = time.Now()

				/* Down Arrow */
			} else if MouseX > int(arrowIconPos.X+(arrowIconSize.X/3)) &&
				MouseX < int(arrowIconPos.X+(arrowIconSize.X/3)*2) &&
				MouseY > int(arrowIconPos.Y+(arrowIconSize.Y/2)) &&
				MouseY < int(arrowIconPos.Y+(arrowIconSize.Y)) {
				dir = DIR_SOUTH
				keyPressed = dir
				//lastKeyPressed = time.Now()

				/* Right Arrow */
			} else if MouseX > int(arrowIconPos.X+(arrowIconSize.X/3)*2) &&
				MouseX < int(arrowIconPos.X+(arrowIconSize.X)) &&
				MouseY > int(arrowIconPos.Y+(arrowIconSize.Y/2)) &&
				MouseY < int(arrowIconPos.Y+(arrowIconSize.Y)) {
				dir = DIR_EAST
				keyPressed = dir
				//lastKeyPressed = time.Now()

			}
		}

		if localPlayer.deadFor == 0 {
			if inpututil.IsKeyJustPressed(ebiten.KeyW) ||
				inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
				dir = DIR_NORTH
			} else if inpututil.IsKeyJustPressed(ebiten.KeyA) ||
				inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
				dir = DIR_WEST
			} else if inpututil.IsKeyJustPressed(ebiten.KeyS) ||
				inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
				dir = DIR_SOUTH
			} else if inpututil.IsKeyJustPressed(ebiten.KeyD) ||
				inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
				dir = DIR_EAST
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			exitLobby()
			return nil
		}

		if dir != 0xFF && GoDirThisTick < 4 &&
			dir != DIR(localPlayer.dir) &&
			dir != reverseDir(localPlayer.dir) {
			sendCommand(CMD_GODIR, uint8ToByteArray(uint8(dir)))
			GoDirThisTick++
			localPlayer.dir = dir
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || TouchPressed {
			if MouseX > ScreenWidth-fsIcon.Bounds().Dx() &&
				MouseY < fsIcon.Bounds().Dy() {
				Fullscreen = ebiten.IsFullscreen()
				ebiten.SetFullscreen(!Fullscreen)
			}
			if verticalMode {
				if MouseX > int(boardPixels)-64 && MouseX < int(boardPixels) &&
					MouseY > int(boardPixels) && MouseY < int(boardPixels)+96 {
					exitLobby()
					return nil
				}
			} else {
				if MouseX > int(boardPixels)+32 && MouseX < int(boardPixels)+96 &&
					MouseY > int(boardPixels-64) && MouseY < int(boardPixels) {
					exitLobby()
					return nil
				}
			}
		}

		/**************
		 * LOBBY MODE *
		 **************/
	} else if gameMode == MODE_LIST_LOBBIES {

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || TouchPressed {
			if MouseY >= cLobbyButtonPosRect.Bounds().Min.Y &&
				MouseY <= cLobbyButtonPosRect.Bounds().Max.Y &&
				MouseX >= cLobbyButtonPosRect.Bounds().Min.X &&
				MouseX <= cLobbyButtonPosRect.Bounds().Max.X {
				sendCommand(CMD_CREATELOBBY, nil)
			}
		}

		if MouseY > lobbiesStartY {
			item := (MouseY-lobbiesStartY)/(int(itemHeight)+lobbySpacing) + scrollOffset
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || TouchPressed {
				clickLobby(item)
			} else {
				showLobby(item)
			}
		} else {
			selectedLobby = -1
		}

		_, y := ebiten.Wheel()
		if y != 0 {
			/* In WASM mode, fix the funky scrolling */
			if WASMMode {
				if time.Since(lastScroll) > time.Millisecond*50 {
					lastScroll = time.Now()
					if y > 0 {
						if scrollOffset > 0 {
							scrollOffset--
						}
					} else {
						if scrollOffset < len(lobbies)-1 {
							scrollOffset++
						}
					}
				}
				/* Otherwise, normal mode */
			} else {

				if y > 0 {
					if scrollOffset > 0 {
						scrollOffset--
					}
				} else {
					if scrollOffset < len(lobbies)-1 {
						scrollOffset++
					}
				}
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyW) ||
			inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
			inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
			if scrollOffset > 0 {
				scrollOffset--
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyS) ||
			inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
			inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
			if scrollOffset < len(lobbies)-1 {
				scrollOffset++
			}
		}

	}

	return nil
}

func clickLobby(item int) {

	lobbyNum := len(lobbies)
	if item >= 0 && item < lobbyNum {
		ID := lobbies[item].ID
		joinLobby(ID)
	}
}

func showLobby(item int) {

	numLobbies := len(lobbies) - 1

	if item <= numLobbies {
		prevSelectedLobby = selectedLobby
		selectedLobby = item
		if prevSelectedLobby != selectedLobby {
			lobbiesDirty = true
		}
	} else {
		selectedLobby = -1
	}

}

func exitLobby() {
	joinLobby(0)
	changeGameMode(MODE_LIST_LOBBIES, 100)
}
