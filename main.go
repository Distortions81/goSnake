package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	_ "net/http/pprof"

	"github.com/hajimehoshi/ebiten/v2"
)

var numPings int

const (
	defaultWindowWidth  = 1280
	defaultWindowHeight = 720
)

func main() {

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		doLog(true, "pprof started")
		pprof.StartCPUProfile(f)
		go func() {
			time.Sleep(time.Minute)
			pprof.StopCPUProfile()
			doLog(true, "pprof complete")
		}()
	}

	/* TODO: use compile flag instead */
	if runtime.GOARCH == "wasm" {
		//doLog(false, "WASM mode")
		WASMMode = true
	}

	/* Load assets */
	loadFonts()
	loadSprites()

	/* Setup ebiten */
	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	/* We manaually clear, so we aren't forced to draw every frame */
	ebiten.SetScreenClearedEveryFrame(false)
	ScreenWidth, ScreenHeight = defaultWindowWidth, defaultWindowHeight

	/* Set up our window */
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("GoSnake")

	doLog(true, "window %v,%v", ScreenWidth, ScreenHeight)

	/* adjust for new window size */
	updateGameSize()

	/*
	 * For testing reasons, connect locally if this isn't a public build
	 * Otherwise, connect to our site and verify everything
	 */

	if buildTime == "Dev" {
		authSite = "https://127.0.0.1/gs"
		transport.TLSClientConfig.InsecureSkipVerify = true
		TouchMode = true //For testing touch control layout
	}

	/* Start ebiten */
	if err := ebiten.RunGameWithOptions(newGame(), nil); err != nil {
		return
	}
}

var (
	transport *http.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client *http.Client = &http.Client{Transport: transport}
)

func newGame() *Game {
	/*
	 * Ebiten started!
	 * Setup sizes for window size
	 * then start our bg loop (async)
	 * and connect to server (async)
	 */

	updateGameSize()
	go bgLooP()
	go checkConnection()
	go connectServer()

	return &Game{}
}

/* Window size chaged, handle it */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	if outsideWidth != ScreenWidth || outsideHeight != ScreenHeight {
		//Lock/Unlock
		gameLock.Lock()
		defer gameLock.Unlock()

		ScreenWidth, ScreenHeight = outsideWidth, outsideHeight
		Fullscreen = ebiten.IsFullscreen()
		updateGameSize()

		/* re-render lobby list */
		if gameMode == MODE_LIST_LOBBIES {
			clearLobbyCache()
			renderLobbyItems()
		}
	}

	return ScreenWidth, ScreenHeight
}

func bgLooP() {

	/* In the background, ping the server or refresh the lobby list */
	go func() {
		for {
			getPing()

			time.Sleep(time.Second)

			if gameMode == MODE_LIST_LOBBIES {
				getLobbies()
			}

			time.Sleep(time.Second * 9)
		}
	}()

}

var gameLagging bool

func checkConnection() {
	for {

		gameLock.Lock()
		if gameMode == MODE_PLAY_GAME {
			if time.Since(lastTickTime) > cTimeout {
				errorText = "Connection timed out."
				changeGameMode(MODE_ERROR, 0)

				gameLock.Unlock()

				connectServer()
				continue
			}
		}
		gameLock.Unlock()

		time.Sleep(time.Second)
	}
}
