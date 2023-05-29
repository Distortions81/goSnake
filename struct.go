package main

type XY struct {
	X int16
	Y int16
}

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
