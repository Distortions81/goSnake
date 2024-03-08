package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const fontDPI = 96.0

func loadFonts() {

	var err error

	/* Mono font */
	fontData := getFont("Hack-Regular.ttf")
	collection, err := opentype.ParseCollection(fontData)
	if err != nil {
		log.Fatal(err)
	}

	mono, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/* Game font */
	fontData = getFont("KomikaAxis.otf")
	collection, err = opentype.ParseCollection(fontData)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * Font DPI
	 * Changes how large the font is for a given point value
	 */

	/* Mono font */
	monoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    8,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Small General font */
	smallGeneralFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    10,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* General font */
	generalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    14,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Large general font */
	largeGeneralFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    20,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Large general font */
	hugeGeneralFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

const sizingText = "!@#$%^&*()_+-=[]{}|;':,.<>?`~qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

func getFontHeight(font font.Face) int {
	tRect := text.BoundString(font, sizingText)
	return tRect.Dy()
}
