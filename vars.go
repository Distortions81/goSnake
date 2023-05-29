package main

import (
	"image/color"
	"sync"
)

const (
	hudSize          = 128
	startGameDelayMS = 1000
	gridSize         = 16
	border           = 1
	tileSize         = gridSize - border
	boardSize        = 50
	gameSpeed        = 4
	deathTicks       = gameSpeed * 2

	DIR_NONE  = 0
	DIR_NORTH = 1
	DIR_EAST  = 2
	DIR_SOUTH = 3
	DIR_WEST  = 4
)

var (
	tiles       map[XY]bool
	players     []playerData
	gameTick    uint64
	gameLock    sync.Mutex
	gameRunning bool
	hudColor    = color.NRGBA{0x20, 0x20, 0x20, 0xff}
	deadColor   = color.NRGBA{0xFF, 0, 0, 0xFF}
)

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
