package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

/* Load all sprites, sub missing ones */
func loadSprites() {
	/* Loads boot screen assets */
	temp, err := getSpriteImage("title-dark.png", false)
	if err == nil {
		titleDark = temp
	}
	temp, err = getSpriteImage("title-dark-blur.png", false)
	if err == nil {
		titleDarkBlur = temp
	}

	temp, err = getSpriteImage("fsIcon.png", false)
	if err == nil {
		fsIcon = temp
	}

	temp, err = getSpriteImage("fsIconM.png", false)
	if err == nil {
		fsIconM = temp
	}

	temp, err = getSpriteImage("apple.png", false)
	if err == nil {
		appleIcon = temp
	}

	temp, err = getSpriteImage("exit.png", false)
	if err == nil {
		exitIcon = temp
	}

	temp, err = getSpriteImage("wasd.png", false)
	if err == nil {
		wasdIcon = temp
	}

	temp, err = getSpriteImage("a-touch.png", false)
	if err == nil {
		touchIcon = temp
	}

	temp, err = getSpriteImage("a-left.png", false)
	if err == nil {
		touchIconLeft = temp
	}

	temp, err = getSpriteImage("a-right.png", false)
	if err == nil {
		touchIconRight = temp
	}

	temp, err = getSpriteImage("a-up.png", false)
	if err == nil {
		touchIconUp = temp
	}

	temp, err = getSpriteImage("a-down.png", false)
	if err == nil {
		touchIconDown = temp
	}

	temp, err = getSpriteImage("icon.png", false)
	if err == nil {
		ebiten.SetWindowIcon([]image.Image{temp})
	}
}
