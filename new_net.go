package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"time"

	"nhooyr.io/websocket"
)

func readNet() {
	if localPlayer == nil ||
		localPlayer.conn == nil ||
		localPlayer.context == nil {
		return
	}

	for {

		_, input, err := localPlayer.conn.Read(localPlayer.context)
		gameLock.Lock()

		if err != nil {
			doLog(true, "readNet error: %v", err)

			errorText = "Connection lost."
			changeGameMode(MODE_ERROR, 0)
			changeGameMode(MODE_BOOT, time.Second)
			gameLock.Unlock()

			connectServer()
			return
		}
		inputLen := len(input)
		if inputLen <= 0 {
			gameLock.Unlock()
			return
		}
		d := CMD(input[0])
		data := input[1:]

		cmdName := cmdNames[d]
		if d != RECV_PLAYERUPDATE {
			if cmdName == "" {
				doLog(true, "Received: 0x%02X (%vb)", d, inputLen)
			} else {
				doLog(true, "Received: %v (%vb)", cmdName, inputLen)
			}
		}

		switch d {
		case CMD_PINGPONG:
			if checkSecret(localPlayer, data) {
				lastRoundTrip = time.Since(pingStart)
				//doLog(true, "PING")
			} else {
				txt := "Authorization failed.\nCheck device:\ntime, date and timezone!"
				errorText = txt
				doLog(true, txt)
				localPlayer.conn.Close(websocket.StatusNormalClosure, "Closed")
				localPlayer.cancel()

				changeGameMode(MODE_ERROR, 0)
				gameLock.Unlock()
				return
			}
		case CMD_JOINLOBBY:
			if gameMode != MODE_LIST_LOBBIES {
				doLog(true, "Received join lobby while not in lobby list mode, ignoring!")
				gameLock.Unlock()
				continue
			}
			gameTick = 0
			lastTickTime = time.Now()
			prevGameTick = 0
			GoDirThisTick = 0
			changeGameMode(MODE_PLAY_GAME, 0)

		case RECV_LOCALPLAYER:
			inBuf := bytes.NewReader(data)
			err = binary.Read(inBuf, binary.LittleEndian, &localPlayer.id)
			if err != nil {
				doLog(true, "%v", err)
				gameLock.Unlock()
				return
			}
			if localPlayer.id == 0 {
				errorText = "Game version too old,\nincorrect date, time or timezone."
				changeGameMode(MODE_ERROR, 0)
				gameLock.Unlock()
				return
			}
		case RECV_LOBBYLIST:
			if gameMode != MODE_LIST_LOBBIES {
				doLog(true, "Received lobby list while not in lobby list mode, ignoring!")
				gameLock.Unlock()
				continue
			}
			unzip := UncompressZip(data)
			err := json.Unmarshal(unzip, &lobbies)

			if err != nil {
				doLog(true, "RECV_LOBBYLIST: Error: %v", err)
				gameLock.Unlock()
				return
			}
			renderLobbyItems()

		case RECV_KEYFRAME:
			if gameMode != MODE_PLAY_GAME && gameMode != MODE_SPECTATE {
				doLog(true, "Received game keyframe while not playing, ignoring!")
				gameLock.Unlock()
				continue
			}
			GoDirThisTick = 0
			gameTick++
			lastTickTime = time.Now()
			deserializeLobbyBinary(data)

		case RECV_PLAYERUPDATE:
			if gameMode != MODE_PLAY_GAME && gameMode != MODE_SPECTATE {
				doLog(true, "Received game update while not playing, ignoring!")
				gameLock.Unlock()
				continue
			}
			GoDirThisTick = 0
			gameTick++
			lastTickTime = time.Now()

			binaryUpdate(data)

		default:
			doLog(true, "Received invalid: 0x%02X\n", d)
			localPlayer.conn.Close(websocket.StatusNormalClosure, "closed")
			gameLock.Unlock()
			return
		}
		gameLock.Unlock()
	}

}

func deserializeLobbyBinary(input []byte) {
	inBuf := bytes.NewReader(input)

	Lobby = lobbyData{}

	var nameLen uint8
	//Lobby Name Len
	err := binary.Read(inBuf, binary.LittleEndian, &nameLen)
	if err != nil {
		doLog(true, "%v", err)
		return
	}

	var lobbyName []byte = make([]byte, nameLen)
	for x := uint8(0); x < nameLen; x++ {
		//Lobby Name Character
		err = binary.Read(inBuf, binary.LittleEndian, &lobbyName[x])
		if err != nil {
			doLog(true, "%v", err)
			return
		}
	}
	Lobby.Name = string(lobbyName)

	//Lobby data
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.ID)
	if err != nil {
		doLog(true, "%v", err)
		return
	}
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.showApple)
	if err != nil {
		doLog(true, "%v", err)
		return
	}
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.apple.X)
	if err != nil {
		doLog(true, "%v", err)
		return
	}
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.apple.Y)
	if err != nil {
		doLog(true, "%v", err)
		return
	}

	//Number of players
	var numPlayers uint16
	err = binary.Read(inBuf, binary.LittleEndian, &numPlayers)
	if err != nil {
		doLog(true, "%v", err)
		return
	}

	for p := 0; p < int(numPlayers); p++ {
		if p >= int(Lobby.NumPlayers) {
			Lobby.players = append(Lobby.players, &playerData{})
			//doLog(true, "Added player")
		}

		var nameLen uint16

		//Player ID
		err = binary.Read(inBuf, binary.LittleEndian, &Lobby.players[p].id)
		if err != nil {
			doLog(true, "%v", err)
			return
		}

		//Player Name Length
		err = binary.Read(inBuf, binary.LittleEndian, &nameLen)
		if err != nil {
			doLog(true, "%v", err)
			return
		}

		var playerName []byte = make([]byte, nameLen)
		for x := uint16(0); x < nameLen; x++ {
			//Player Name Character
			err = binary.Read(inBuf, binary.LittleEndian, &playerName[x])
			if err != nil {
				doLog(true, "%v", err)
				return
			}
		}
		Lobby.players[p].Name = string(playerName)

		//Player Dead For
		err = binary.Read(inBuf, binary.LittleEndian, &Lobby.players[p].deadFor)
		if err != nil {
			doLog(true, "%v", err)
			return
		}
		if Lobby.players[p].id == localPlayer.id {
			localPlayer.deadFor = Lobby.players[p].deadFor
		}

		//Player Length
		err = binary.Read(inBuf, binary.LittleEndian, &Lobby.players[p].length)
		if err != nil {
			doLog(true, "%v", err)
			return
		}

		Lobby.players[p].tiles = nil

		for x := uint16(0); x < Lobby.players[p].length; x++ {
			var tileX, tileY uint8

			//Tile position
			err = binary.Read(inBuf, binary.LittleEndian, &tileX)
			if err != nil {
				doLog(true, "%v", err)
				return
			}
			err = binary.Read(inBuf, binary.LittleEndian, &tileY)
			if err != nil {
				doLog(true, "%v", err)
				return
			}

			if x >= uint16(len(Lobby.players[p].tiles)) {
				Lobby.players[p].tiles = append(Lobby.players[p].tiles, XY{})
				//doLog(true, "Add tile")
			}
			Lobby.players[p].tiles[x].X = tileX
			Lobby.players[p].tiles[x].Y = tileY
		}
	}
	Lobby.NumPlayers = numPlayers
}

// This can be further optimized, once game logic is put into a module both use.
func binaryUpdate(input []byte) {
	inBuf := bytes.NewReader(input)

	//Apple position
	err := binary.Read(inBuf, binary.LittleEndian, &Lobby.showApple)
	if err != nil {
		doLog(true, "%v", err)
		return
	}
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.apple.X)
	if err != nil {
		doLog(true, "%v", err)
		return
	}
	err = binary.Read(inBuf, binary.LittleEndian, &Lobby.apple.Y)
	if err != nil {
		doLog(true, "%v", err)
		return
	}

	//Number of players
	var numPlayers uint16
	err = binary.Read(inBuf, binary.LittleEndian, &numPlayers)
	if err != nil {
		doLog(true, "%v", err)
		return
	}

	for p := 0; p < int(numPlayers); p++ {
		if p >= int(Lobby.NumPlayers) {
			Lobby.players = append(Lobby.players, &playerData{})
			errorText = "Extra player data, desync!"
			changeGameMode(MODE_ERROR, 1000)
		}

		//Player Dead For
		err = binary.Read(inBuf, binary.LittleEndian, &Lobby.players[p].deadFor)
		if err != nil {
			doLog(true, "%v", err)
			return
		}
		if Lobby.players[p].id == localPlayer.id {
			localPlayer.deadFor = Lobby.players[p].deadFor
		}

		//Player Length
		err = binary.Read(inBuf, binary.LittleEndian, &Lobby.players[p].length)
		if err != nil {
			doLog(true, "%v", err)
			return
		}

		Lobby.players[p].tiles = nil

		for x := uint16(0); x < Lobby.players[p].length; x++ {
			var tileX, tileY uint8

			//Tile position
			err = binary.Read(inBuf, binary.LittleEndian, &tileX)
			if err != nil {
				doLog(true, "%v", err)
				return
			}
			err = binary.Read(inBuf, binary.LittleEndian, &tileY)
			if err != nil {
				doLog(true, "%v", err)
				return
			}

			if x >= uint16(len(Lobby.players[p].tiles)) {
				Lobby.players[p].tiles = append(Lobby.players[p].tiles, XY{})
				//doLog(true, "Add tile")
			}
			Lobby.players[p].tiles[x].X = tileX
			Lobby.players[p].tiles[x].Y = tileY
		}
	}
	Lobby.NumPlayers = numPlayers
}

func sendCommand(header CMD, data []byte) bool {
	if localPlayer == nil || localPlayer.context == nil || localPlayer.conn == nil {
		return false
	}

	cmdName := cmdNames[header]
	if cmdName == "" {
		doLog(true, "Sent: 0x%02X", header)
	} else {
		doLog(true, "Sent: %v", cmdName)
	}

	var err error
	if data == nil {
		err = localPlayer.conn.Write(localPlayer.context, websocket.MessageBinary, []byte{byte(header)})
	} else {
		err = localPlayer.conn.Write(localPlayer.context, websocket.MessageBinary, append([]byte{byte(header)}, data...))
	}
	if err != nil {

		doLog(true, "sendCommand error: %v", err)

		errorText = "Connection lost."
		changeGameMode(MODE_ERROR, 0)
		changeGameMode(MODE_BOOT, time.Second)
		connectServer()
		return false
	}

	return true
}
