package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"math"
	"sort"
	"time"
)

/* Generic unzip []byte */
func UncompressZip(data []byte) []byte {
	b := bytes.NewReader(data)

	z, _ := zlib.NewReader(b)
	defer z.Close()

	p, err := io.ReadAll(z)
	if err != nil {
		return nil
	}
	return p
}

/* Generic zip []byte */
func CompressZip(data []byte) []byte {
	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, zlib.BestCompression)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func uint64ToByteArray(i uint64) []byte {
	byteArray := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteArray, i)
	return byteArray
}

func uint32ToByteArray(i uint32) []byte {
	byteArray := make([]byte, 4)
	binary.LittleEndian.PutUint32(byteArray, i)
	return byteArray
}

func uint16ToByteArray(i uint16) []byte {
	byteArray := make([]byte, 2)
	binary.LittleEndian.PutUint16(byteArray, i)
	return byteArray
}

func uint8ToByteArray(i uint8) []byte {
	byteArray := make([]byte, 1)
	byteArray[0] = byte(i)
	return byteArray
}

func byteArrayToUint8(i []byte) uint8 {
	if len(i) < 1 {
		return 0
	}
	return uint8(i[0])
}

func byteArrayToUint16(i []byte) uint16 {
	if len(i) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(i)
}

func byteArrayToUint32(i []byte) uint32 {
	if len(i) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(i)
}

func byteArrayToUint64(i []byte) uint64 {
	if len(i) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(i)
}

func sortByScore(players []*playerData) []*playerData {

	var sorted []*playerData
	for p := range players {
		sorted = append(sorted, players[p])
		sorted[p].lid = p
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].length > sorted[j].length
	})
	return sorted

}

func distance(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

func midPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

func changeGameMode(newMode MODE, delay time.Duration) {

	/* Skip if the same */
	if newMode == gameMode {
		return
	}

	/* If we were in lobby mode, but changed mode... cleanup */
	if gameMode == MODE_LIST_LOBBIES {
		clearLobbyCache()

		/* If we are going to lobby mode, prep */
	} else if newMode == MODE_LIST_LOBBIES {
		prevScrollOffset = 0
		scrollOffset = 1

		numPings = 0
		time.Sleep(time.Millisecond * 100)
		getPing()
		time.Sleep(time.Millisecond * 100)
		getLobbies()
	} else if newMode == MODE_PLAY_GAME {
		localPlayer.tiles = nil
		localPlayer.length = 0
		localPlayer.dir = 0xFF
		Lobby = lobbyData{}
	}

	time.Sleep(delay)

	gameMode = newMode
}

func reverseDir(dir DIR) DIR {
	switch dir {
	case DIR_NORTH:
		return DIR_SOUTH
	case DIR_EAST:
		return DIR_WEST
	case DIR_SOUTH:
		return DIR_NORTH
	case DIR_WEST:
		return DIR_EAST
	}
	return dir
}

func clearLobbyCache() {

	numPings = 0

	count := 0
	for l := range lobbies {
		count++
		if lobbies[l].image == nil {
			continue
		}
		lobbies[l].image.Dispose()
		lobbies[l].image = nil
	}

	/* If lobby list shrinks, scroll back */
	if count < scrollOffset {
		scrollOffset = count
	}
}

func PosIntMod(d, m int) int {
	var res int = d % m
	if res < 0 && m > 0 {
		return res + m
	}
	return res
}
