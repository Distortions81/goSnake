package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

var (
	/* Build data */
	WASMMode  bool
	buildTime string = "Dev"

	/* Screen */
	ScreenWidth, ScreenHeight int
	Fullscreen, verticalMode  bool

	/* Mouse */
	lastMouseX, lastMouseY,
	MouseX, MouseY int

	/* Game mode */
	gameMode  = MODE_START
	errorText string

	/* Local player */
	localPlayer *playerData
	keyPressed  = DIR_NONE

	/* Lobbies */
	lobbies           []lobbyData
	lobbiesDirty      bool = true
	Lobby             lobbyData
	selectedLobby     int
	prevSelectedLobby int

	/* Networking */
	authSite = "https://gosnake.go-game.net/gs"

	/* Fonts */
	monoFont,
	smallGeneralFont,
	generalFont,
	largeGeneralFont,
	hugeGeneralFont font.Face

	/* Images */

	titleDark,
	titleDarkBlur,
	appleIcon,
	fsIcon,
	fsIconM,
	exitIcon,
	touchIcon,
	touchIconLeft,
	touchIconRight,
	touchIconUp,
	touchIconDown,
	wasdIcon *ebiten.Image

	/* Ping */
	statusTime    time.Time
	lastRoundTrip time.Duration

	/* Reconnect */
	ReconnectCount     = 0
	RecconnectDelayCap = 30

	/* Game board values */
	boardPixels         = boardSize * gridSize
	gridSize    uint16  = uint16(ScreenHeight / boardSize)
	tileSize    uint16  = gridSize - tileBorder
	halfGrid    float32 = float32(gridSize) / 2.0

	/* WASM Scroll fix */
	scrollOffset         = 0
	prevScrollOffset     = 0
	scrollDown       int = 0
	scrollUp         int = 0

	numColors = len(colorList)
)

var colorList = []color.NRGBA{
	{255, 255, 255, 255},
	{40, 180, 99, 255},
	{41, 128, 185, 255},
	{244, 208, 63, 255},
	{243, 156, 18, 255},
	{255, 151, 197, 255},
	{165, 105, 189, 255},
	{209, 209, 209, 255},
	{64, 199, 178, 255},
	{99, 114, 166, 255},
	{134, 166, 99, 255},
	{206, 231, 114, 255},
	{209, 114, 231, 255},
	{114, 228, 231, 255},
	{176, 116, 78, 255},
	{210, 113, 52, 255},
}
