package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"nhooyr.io/websocket"
)

const netReadLimit = 1024 * 1024 * 10 //10mb

func connectServer() {

	changeGameMode(MODE_BOOT, 0)

	if localPlayer != nil {
		if localPlayer.conn != nil {
			localPlayer.cancel()
			localPlayer.conn.Close(websocket.StatusNormalClosure, "Write failed.")
		}

		localPlayer = nil
	}

	for !doConnect() {
		ReconnectCount++
		offset := ReconnectCount
		if offset > RecconnectDelayCap {
			offset = RecconnectDelayCap
		}
		timeFuzz := rand.Int63n(200000000) //2 seconds of random delay
		delay := float64(3 + offset + int(float64(timeFuzz)/100000000.0))
		statusTime = time.Now().Add(time.Duration(delay) * time.Second)

		changeGameMode(MODE_RECONNECT, time.Millisecond*500)
		doLog(true, "Connect %v failed, retrying in %v ...", ReconnectCount, time.Until(statusTime).Round(time.Second).String())
		time.Sleep(time.Duration(delay) * time.Second)
	}
	time.Sleep(time.Millisecond * 500)
	changeGameMode(MODE_LIST_LOBBIES, 0)
}

func doConnect() bool {

	changeGameMode(MODE_CONNECT, 0)
	doLog(true, "Connecting...")

	ctx, cancel := context.WithCancel(context.Background())

	c, err := platformDial(ctx)

	if err != nil {
		log.Printf("dial failed: %v\n", err)
		cancel()
		return false
	}

	//10MB limit
	c.SetReadLimit(netReadLimit)

	localPlayer = &playerData{conn: c, context: ctx, cancel: cancel, id: 0}

	secret := generateSecret(nil)
	c.Write(ctx, websocket.MessageBinary, append([]byte{byte(CMD_INIT)}, secret...))
	doLog(true, "Connected!")

	changeGameMode(MODE_CONNECTED, 0)
	go readNet()

	return true
}

var pingStart time.Time

func getPing() {
	pingStart = time.Now()
	sendCommand(CMD_PINGPONG, generateSecret(localPlayer))
}

func getLobbies() bool {
	sendCommand(CMD_GETLOBBIES, nil)
	return true
}

func joinLobby(id uint16) bool {
	Lobby = lobbyData{}
	localPlayer.tiles = []XY{}
	sendCommand(CMD_JOINLOBBY, uint16ToByteArray(id))
	return true
}
