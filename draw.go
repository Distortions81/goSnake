package main

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

const (
	/* Game tiles */
	tileBorder = 1  //1px spacing
	boardSize  = 32 //Default board size

	/* Lobby drawing */
	lobbySpacing           = 8
	lobbiesStartX          = 4
	lobbiesStartY          = 64
	lobbyTextOffsetX       = 8
	lobbyTextOffsetY       = 8
	lobbyWidthSpacing      = 4
	lobbyPlayerNameMax     = 18
	lobbyNameMax           = 128
	lobbyAllPlayerNamesMax = 256
	makeLobbyText          = "Create New Lobby"
	boundString            = " !\"#$%&\\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

	lineSize   float32 = 36.0
	itemHeight float32 = lineSize * 2
	lineThick  float32 = 2.0
)

var (
	gameTick       uint64
	prevGameTick   uint64
	lastTickTime   time.Time
	lastLobbyFrame time.Time
)

/* Optimize me */
func (g *Game) Draw(screen *ebiten.Image) {

	startTime := time.Now()

	//Lock/Unlock
	gameLock.Lock()
	defer gameLock.Unlock()

	tColor := color.White

	if gameMode == MODE_ERROR {
		screen.Clear()
		drawBG(screen, false)
		drawText(errorText, largeGeneralFont, color.White, ColorSmokedGlass, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 8, screen, false, true, true)

	} else if gameMode == MODE_START {
		screen.Clear()
		drawBG(screen, false)

	} else if gameMode == MODE_BOOT {
		screen.Clear()
		drawBG(screen, false)
		drawText("GoSnake", hugeGeneralFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 0, screen, false, true, true)

	} else if gameMode == MODE_CONNECT {
		screen.Clear()
		drawBG(screen, false)
		drawText("Connecting...", hugeGeneralFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 0, screen, false, true, true)

	} else if gameMode == MODE_RECONNECT {
		screen.Clear()
		drawBG(screen, false)
		buf := "Retrying..."
		if time.Until(statusTime) > time.Second {
			buf = fmt.Sprintf("Connection failed.\nRetrying in %s ...", time.Until(statusTime).Round(time.Second).String())
		}
		drawText(buf, largeGeneralFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 0, screen, false, true, true)

	} else if gameMode == MODE_CONNECTED {
		screen.Clear()
		drawBG(screen, false)
		drawText("Getting lobbies...", largeGeneralFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 0, screen, false, true, true)

	} else if gameMode == MODE_SELECT_LOBBY {
		screen.Clear()

		drawBG(screen, false)
		drawText("Joining lobby...", largeGeneralFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: uint16((ScreenHeight) / 4.0)}, 0, screen, false, true, true)

	} else if gameMode == MODE_LIST_LOBBIES {
		/* If we scrolled, the lobby list was updated, selection changed or 30fps min */
		if scrollOffset != prevScrollOffset ||
			lobbiesDirty ||
			time.Since(lastLobbyFrame) < time.Millisecond*33 {

			lobbiesDirty = false
			lastLobbyFrame = time.Now()

			screen.Clear()
			drawLobbyList(screen, tColor)
			drawDebugInfo(screen)
		}
	} else if gameMode == MODE_PLAY_GAME {
		if gameTick != prevGameTick {
			prevGameTick = gameTick
			screen.Clear()
			drawGame(screen)
			drawDebugInfo(screen)
		}
	}

	took := time.Since(startTime)

	if !WASMMode {
		/* Cap to 500fps */
		if took < time.Millisecond*2 {
			time.Sleep((time.Millisecond * 2) - took)
		}
	}

}

const fpad = 4

func drawFsIcon(screen *ebiten.Image) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(ScreenWidth)-(float64(fsIcon.Bounds().Dy()+fpad)), fpad)
	if Fullscreen {
		screen.DrawImage(fsIconM, opt)
	} else {
		screen.DrawImage(fsIcon, opt)
	}
}

var cLobbyButtonPosRect image.Rectangle

func drawLobbyList(screen *ebiten.Image, tColor color.Color) {
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	drawBG(screen, true)

	buf := fmt.Sprintf("Lobbies: %v, Ping: %v\n\n", len(lobbies), lastRoundTrip.Round(time.Millisecond).String())
	drawText(buf, generalFont, tColor, color.Transparent, XY16{X: uint16((ScreenWidth) / 2.0), Y: 55}, 0, screen, false, true, true)
	tPos := XY16{X: uint16(ScreenWidth) - 10, Y: 10}
	cLobbyButtonPosRect = rectDrawText(makeLobbyText, generalFont, color.Black, ColorRed, tPos, 2, screen, false, false, false)

	op.GeoM.Translate(lobbiesStartX, lobbiesStartY)
	if scrollOffset > prevScrollOffset {
		scrollUp = 2

	} else if scrollOffset < prevScrollOffset {
		scrollDown = 2

	}
	if scrollUp > 0 {
		op.GeoM.Translate(0, float64(itemHeight/4)*float64(scrollUp))
		scrollUp--
	} else if scrollDown > 0 {
		op.GeoM.Translate(0, -float64(itemHeight/4)*float64(float64(scrollDown)))
		scrollDown--
	}

	lobNum := len(lobbies)
	spacing := float64(itemHeight + lobbySpacing)
	vertSpace := ScreenHeight/int(spacing) - 1

	for x := scrollOffset; x < scrollOffset+vertSpace; x++ {
		op.ColorScale.Reset()

		if x < 0 || x >= lobNum {
			op.GeoM.Translate(0, spacing)

			continue
		}
		if lobbies[x].image == nil {
			continue
		}
		if x == selectedLobby {
			op.ColorScale.ScaleAlpha(0.66)
		}

		screen.DrawImage(lobbies[x].image, op)
		op.GeoM.Translate(0, spacing)
	}

	/*
		if toolTipString != "" {
			drawText(toolTipString, monoFont, color.White, ColorToolTipBG,
				XY{X: uint16(MouseX + 32), Y: uint16(MouseY) + 32},
				4, screen, true, false, false)
		}
	*/

	prevScrollOffset = scrollOffset
}

func drawGame(screen *ebiten.Image) {
	drawBG(screen, true)

	/* Draw game back */
	vector.DrawFilledRect(screen, float32(halfGrid), float32(halfGrid), float32(boardSize*gridSize), float32(boardSize*gridSize), ColorSmokedGlass, false)

	/* Draw Apple */
	if Lobby.showApple {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

		spriteSize := appleIcon.Bounds().Dx()
		scale := 1.0 / float32(spriteSize/int(gridSize))
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(uint16(Lobby.apple.X-1)*gridSize)+float64(halfGrid), float64(uint16(Lobby.apple.Y-1)*gridSize)+float64(halfGrid))

		vector.DrawFilledRect(screen, float32(uint16(Lobby.apple.X-1)*gridSize)+halfGrid, float32(uint16(Lobby.apple.Y-1)*gridSize)+halfGrid, float32(tileSize), float32(tileSize), ColorMilkGlass, false)
		screen.DrawImage(appleIcon, op)
	}

	for p, player := range Lobby.players {
		if player.length <= 0 || player.deadFor > 8 || player.deadFor < -8 {
			continue
		}
		for _, tile := range player.tiles {
			if player.deadFor != 0 && gameTick%2 == 0 {
				vector.DrawFilledRect(screen, float32(uint16(tile.X-1)*gridSize)+halfGrid, float32(uint16(tile.Y-1)*gridSize)+halfGrid, float32(tileSize), float32(tileSize), ColorSmokedGlass, false)
			} else if player.id == localPlayer.id {
				vector.DrawFilledCircle(screen, float32(uint16(tile.X-1)*gridSize)+(halfGrid*2), float32(uint16(tile.Y-1)*gridSize)+(halfGrid*2), float32(tileSize/2), ColorRed, true)
			} else {
				vector.DrawFilledRect(screen, float32(uint16(tile.X-1)*gridSize)+halfGrid, float32(uint16(tile.Y-1)*gridSize)+halfGrid, float32(tileSize), float32(tileSize), colorList[PosIntMod(p, numColors)], false)
			}
		}
	}

	fsize := uint16(getFontHeight(smallGeneralFont))
	var row uint16 = 2

	scoreList := sortByScore(Lobby.players)

	count := 0
	for _, test := range Lobby.players {
		if test.deadFor == 0 {
			count++
		}
	}
	buf := fmt.Sprintf("Lobby: '%v' players: (%v)", Lobby.Name, count)

	/* Show touchscreen response */
	keyImage := touchIcon
	if TouchMode {
		if keyPressed == DIR_NORTH {
			keyImage = touchIconUp
			keyPressed = DIR_NONE
		} else if keyPressed == DIR_EAST {
			keyImage = touchIconRight
			keyPressed = DIR_NONE
		} else if keyPressed == DIR_SOUTH {
			keyImage = touchIconDown
			keyPressed = DIR_NONE
		} else if keyPressed == DIR_WEST {
			keyImage = touchIconLeft
			keyPressed = DIR_NONE
		}
	}

	/*
	 * VERTICAL MODE
	 */
	if verticalMode {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

		spriteSize := exitIcon.Bounds().Dx()
		scale := 1.0 / float32(spriteSize/int(64))
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(boardPixels-64), float64(boardPixels+32))
		screen.DrawImage(exitIcon, op)

		drawText(buf, smallGeneralFont, color.White, ColorSmokedGlass, XY16{Y: boardPixels + gridSize, X: gridSize}, 2, screen, true, false, false)
		for _, pitem := range scoreList {
			if pitem.length > 0 {
				buf := fmt.Sprintf("%v: %v", pitem.Name, pitem.length)
				pcolor := colorList[PosIntMod(pitem.lid, numColors)]
				if pitem.id == localPlayer.id {
					buf = buf + " (you)"
					pcolor = ColorRed
				}
				yp := (fsize * row) + boardPixels + gridSize
				if yp+fsize < uint16(ScreenHeight) {
					drawText(buf, smallGeneralFont, pcolor, color.Transparent, XY16{X: gridSize, Y: yp}, 2, screen, true, false, false)
				}
				row++
			}
		}

		/* NON TOUCH MODE */
		if !TouchMode {
			op.GeoM.Reset()
			spriteSize = wasdIcon.Bounds().Dx()
			scale = 1.0 / float32(spriteSize/int(64))
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(float64(boardPixels-150), float64(boardPixels+32))
			screen.DrawImage(wasdIcon, op)
		} else {
			/* TOUCH MODE */
			op.GeoM.Reset()
			op.GeoM.Scale(float64(arrowScale), float64(arrowScale))
			op.GeoM.Translate(float64(arrowIconPos.X), float64(arrowIconPos.Y))
			screen.DrawImage(keyImage, op)
		}
	} else {
		/*
		 * HORIZONTAL MODE
		 */
		if !WASMMode {
			drawFsIcon(screen)
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

		spriteSize := exitIcon.Bounds().Dx()
		scale := 1.0 / float32(spriteSize/int(64))
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(boardPixels+32), float64(boardPixels-64))
		screen.DrawImage(exitIcon, op)

		drawText(buf, smallGeneralFont, color.White, ColorSmokedGlass, XY16{X: boardPixels + gridSize, Y: fsize}, 2, screen, true, false, false)
		for _, pitem := range scoreList {
			if pitem.length > 0 {
				row++
				buf := fmt.Sprintf("%v: %v", pitem.Name, pitem.length)
				pcolor := colorList[PosIntMod(pitem.lid, numColors)]
				if pitem.id == localPlayer.id {
					buf = buf + " (you)"
					pcolor = ColorRed
				}
				yp := fsize * row
				if yp+fsize < uint16(ScreenHeight) {
					drawText(buf, smallGeneralFont, pcolor, color.Transparent, XY16{X: boardPixels + gridSize, Y: yp}, 2, screen, true, false, false)
				}
			}
		}

		/* NON TOUCH MODE */
		if !TouchMode {
			op.GeoM.Reset()
			spriteSize = wasdIcon.Bounds().Dx()
			scale = 1.0 / float32(spriteSize/int(64))
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(float64(boardPixels+32), float64(boardPixels-128))
			screen.DrawImage(wasdIcon, op)
		} else {
			/* TOUCH MODE */
			op.GeoM.Reset()
			op.GeoM.Scale(float64(arrowScale), float64(arrowScale))
			op.GeoM.Translate(float64(arrowIconPos.X), float64(arrowIconPos.Y))
			screen.DrawImage(keyImage, op)
		}
	}

}

/* Optimize me */
func drawBG(screen *ebiten.Image, blur bool) {
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	var img *ebiten.Image
	if blur {
		img = titleDarkBlur
	} else {
		img = titleDark
	}
	spriteBounds := img.Bounds()
	scalex := 1.0 / (float64(spriteBounds.Size().X) / float64(ScreenWidth))
	scaley := 1.0 / (float64(spriteBounds.Size().Y) / float64(ScreenHeight))
	if scalex > scaley {
		op.GeoM.Scale(scalex, scalex)
	} else {
		op.GeoM.Scale(scaley, scaley)
	}

	op.GeoM.Translate(
		float64(ScreenWidth/2)-(float64(spriteBounds.Dx()/2)*scalex),
		float64(ScreenHeight/2)-(float64(spriteBounds.Dy()/2)*scaley),
	)

	//screen.Fill(ColorNearBlack)
	screen.DrawImage(img, op)
}

func renderLobbyItems() {

	clearLobbyCache()
	lobbiesDirty = true

	for l, lobby := range lobbies {

		nPrefix := ""
		if lobby.NumPlayers > 0 {
			nPrefix = fmt.Sprintf("%v", lobby.NumPlayers)
		} else {
			nPrefix = "None"
		}

		buf := fmt.Sprintf("%v\nPlayers (%v): %v", TruncateStringEllipsis(lobby.Name, lobbyNameMax), nPrefix, lobby.PlayerNames)
		pos := XY16{X: lobbyTextOffsetX, Y: uint16(itemHeight + lobbyTextOffsetY)}
		xWidth := float32(ScreenWidth - lobbyWidthSpacing)

		if lobbies[l].image == nil {
			lobbies[l].image = ebiten.NewImage(int(xWidth), int(itemHeight))
		}

		borderColor := ColorVeryDarkRed

		//BG
		vector.DrawFilledRect(lobbies[l].image, 0, 0, xWidth, itemHeight, ColorSmokedGlass, false)
		//Text
		rectDrawText(buf, generalFont, color.White, color.Transparent, pos, 0, lobbies[l].image, true, true, false)
		//Left
		vector.DrawFilledRect(lobbies[l].image, 0, 0, lineThick, itemHeight, borderColor, false)
		//Right
		vector.DrawFilledRect(lobbies[l].image, xWidth-lineThick, 0, lineThick, itemHeight, borderColor, false)
		//Top
		vector.DrawFilledRect(lobbies[l].image, 0, 0, xWidth, lineThick, borderColor, false)
		//Bottom
		vector.DrawFilledRect(lobbies[l].image, 0, itemHeight-lineThick, xWidth, lineThick, borderColor, false)
	}
}

var updateGameSizeLock sync.Mutex

func updateGameSize() {

	/* Resize everything for the new window size */
	updateGameSizeLock.Lock()
	defer updateGameSizeLock.Unlock()

	if ScreenWidth > ScreenHeight {
		gridSize = uint16(ScreenHeight / (boardSize + 1))
		if gridSize < 3 {
			gridSize = 3
		}
		tileSize = gridSize - tileBorder

		boardPixels = boardSize * gridSize
		if boardPixels < (boardSize+1)*3 {
			boardPixels = (boardSize + 1) * 3
		}
		halfGrid = float32(gridSize) / 2.0

		verticalMode = false
	} else {
		gridSize = uint16(ScreenWidth / (boardSize + 1))
		if gridSize < 3 {
			gridSize = 3
		}
		tileSize = gridSize - tileBorder

		boardPixels = boardSize * gridSize
		if boardPixels < (boardSize+1)*3 {
			boardPixels = (boardSize + 1) * 3
		}
		halfGrid = float32(gridSize) / 2.0

		verticalMode = true
	}

	updateTouchControls()
}

func drawDebugInfo(screen *ebiten.Image) {

	/* Draw debug info */
	buf := fmt.Sprintf("FPS: %-4v Arch: %v Build: v%v-%v Ping: %v",
		int(ebiten.ActualFPS()),
		runtime.GOARCH, protoVersion, buildTime,
		lastRoundTrip.Round(time.Microsecond).String(),
	)

	drawText(buf, monoFont, ColorDebugFG, ColorDebugBG,
		XY16{X: 0, Y: uint16(ScreenHeight) + 10},
		1, screen, true, true, false)

}

func rectDrawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos XY16,
	pad int, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) image.Rectangle {
	var tmx, tmy float32

	tRect := text.BoundString(face, input)

	if justCenter {
		tmx = float32(int(pos.X) - (tRect.Dx() / 2))
		tmy = float32(int(pos.Y) - (tRect.Dy() / 2))
	} else {
		if justLeft {
			tmx = float32(pos.X)
		} else {
			tmx = float32(int(pos.X) - tRect.Dx())
		}

		if justUp {
			tmy = float32(int(pos.Y) - tRect.Dy())
		} else {
			tmy = float32(pos.Y) + float32(tRect.Dy())
		}
	}

	fHeight := text.BoundString(face, boundString)

	xPos := tmx - float32(pad)
	yPos := tmy - float32(fHeight.Dy()/2) - float32(pad)
	xWidth := float32(tRect.Dx()) + float32(pad*2)
	yWidth := float32(tRect.Dy()) + float32(pad*2)

	if screen != nil {
		vector.DrawFilledRect(
			screen, xPos, yPos,
			xWidth, yWidth, bgcolor, false,
		)
		text.Draw(screen, input, face, int(tmx), int(tmy), color)
	}

	result := image.Rectangle{}
	result.Min.X = int(xPos)
	result.Min.Y = int(yPos)

	result.Max.X = int(xPos + xWidth)
	result.Max.Y = int(yPos + yWidth)
	return result

}

func drawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos XY16,
	pad int, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) XY16 {

	rrect := rectDrawText(input, face, color, bgcolor, pos, pad, screen, justLeft, justUp, justCenter)
	return XY16{X: uint16(rrect.Dx()), Y: uint16(rrect.Dy())}
}
