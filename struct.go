package main

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"
	"nhooyr.io/websocket"
)

type lobbyData struct {
	ID   uint16 `json:"i"`
	Name string `json:"n"`

	players     []*playerData
	PlayerNames string `json:"p"`
	NumPlayers  uint16 `json:"c"`

	showApple bool
	apple     XY

	image *ebiten.Image
}

type playerData struct {
	conn    *websocket.Conn
	context context.Context
	cancel  context.CancelFunc

	lid int

	id   uint32
	Name string
	dir  DIR

	deadFor int8
	length  uint16

	tiles []XY
}

type XY struct {
	X uint8
	Y uint8
}

type XY16 struct {
	X uint16
	Y uint16
}

type Game struct {
}
