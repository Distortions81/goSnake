package main

import (
	"bytes"
	"embed"
	"image"
	"io"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	//go:embed data
	f embed.FS
)

const cLoadEmbedSprites = true

func getFont(name string) []byte {
	data, err := f.ReadFile(gfxDir + "fonts/" + name)
	if err != nil {
		log.Fatal(err)
	}
	return data

}

func getSpriteImage(name string, unmanaged bool) (*ebiten.Image, error) {

	if cLoadEmbedSprites {
		gpng, err := f.Open(gfxDir + name)
		if err != nil {
			doLog(true, "GetSpriteImage: Open: %v", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			doLog(true, "GetSpriteImage: Decode: %v", err)
			return nil, err
		}
		var img *ebiten.Image
		if unmanaged {
			img = ebiten.NewImageFromImageWithOptions(m, &ebiten.NewImageFromImageOptions{Unmanaged: true})
		} else {
			img = ebiten.NewImageFromImage(m)
		}
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile(dataDir + gfxDir + name)
		if err != nil {
			doLog(true, "GetSpriteImage: File: %v", err)
		}
		return img, err
	}
}

var audioPlayer *audio.Player

func playMusic(name string) {
	if audioPlayer != nil {
		return
	}
	doLog(true, "Loading music...")

	/* Test music */
	sampleRate := 48000

	/* Fetch data */
	musicBytes, err := getMusicBytes("title")
	if err != nil {
		doLog(true, "playMusic: %v", err)
		return
	}

	/* Create context */
	audioCon := audio.NewContext(sampleRate)

	/* Decode MP3 */
	mp3Data, err := mp3.DecodeWithoutResampling(bytes.NewReader(musicBytes))
	if err != nil {
		doLog(true, "playMusic: %v", err)
		return
	}

	/* Read all mp3 data*/
	audioData, err := io.ReadAll(mp3Data)
	if err != nil {
		doLog(true, "playMusic: %v", err)
		return
	}

	/* Create player */
	audioPlayer = audioCon.NewPlayerFromBytes(audioData)

	doLog(true, "Music ready...")
}

func getMusicBytes(name string) ([]byte, error) {

	df, err := f.Open(musicDir + name + ".mp3")
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(df)
	if err != nil {
		return nil, err
	}
	return data, err
}

func getText(name string) (string, error) {
	file, err := f.Open(txtDir + name + ".txt")
	if err != nil {
		doLog(true, "GetText: %v", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		doLog(true, "GetText: %v", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		doLog(true, "GetText: %v", name)
		return strings.ReplaceAll(string(txt), "\r", ""), nil
	} else {
		return "Error: length 0!", err
	}

}
