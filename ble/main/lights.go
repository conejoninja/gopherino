package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	numLEDs = 4

	flashTime = 250 * time.Millisecond
	blinkTime = 150 * time.Millisecond // Auto mode, while dodging
)

var (
	neo   machine.Pin = machine.P15
	ws    ws2812.Device
	leds  [numLEDs]color.RGBA
	shown [numLEDs]color.RGBA

	// The two small red LEDs at the front, used as momentary "turn
	// signals" by flashLeft / flashRight.
	ledLeft  machine.Pin = machine.P8
	ledRight machine.Pin = machine.P12

	flashAllUntil, flashLeftUntil, flashRightUntil time.Time

	colorOff   = color.RGBA{0, 0, 0, 255}
	colorBlue  = color.RGBA{0, 0, 160, 255}
	colorGreen = color.RGBA{0, 160, 0, 255}
	colorRed   = color.RGBA{160, 0, 0, 255}
	colorWhite = color.RGBA{160, 160, 160, 255}
)

func initLights() {
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws = ws2812.NewWS2812(neo)
	ledLeft.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledRight.Configure(machine.PinConfig{Mode: machine.PinOutput})
	fill(colorOff)
	ws.WriteColors(leds[:])
}

func flash(which byte, now time.Time) {
	switch which {
	case flashAll:
		flashAllUntil = now.Add(flashTime)
	case flashLeft:
		flashLeftUntil = now.Add(flashTime)
	case flashRight:
		flashRightUntil = now.Add(flashTime)
	}
}

func resetFlashes() {
	flashAllUntil = time.Time{}
	flashLeftUntil = time.Time{}
	flashRightUntil = time.Time{}
}

// updateLights paints the current mode's colors (plus any momentary flash)
// and only talks to the LEDs when something changed.
func updateLights(now time.Time) {
	switch mode {
	case modeNormal:
		fill(colorBlue)
	case modeTank:
		fill(colorGreen)
	case modeMusic:
		rainbow(now)
	case modeAuto:
		if obstacleActive() && (now.UnixMilli()/blinkTime.Milliseconds())%2 == 0 {
			fill(colorOff)
		} else {
			fill(colorRed)
		}
	}
	if now.Before(flashAllUntil) {
		fill(colorWhite)
	}
	ledLeft.Set(now.Before(flashLeftUntil))
	ledRight.Set(now.Before(flashRightUntil))

	if leds != shown {
		shown = leds
		ws.WriteColors(leds[:])
	}
}

func fill(c color.RGBA) {
	for i := range leds {
		leds[i] = c
	}
}

// rainbow spreads the color wheel across the LEDs and slowly rotates it.
func rainbow(now time.Time) {
	base := int(now.UnixMilli()/8) % 256
	for i := range leds {
		leds[i] = wheel(byte(base + i*256/numLEDs))
	}
}

// wheel maps 0..255 to a color going red -> green -> blue -> red.
func wheel(pos byte) color.RGBA {
	const max = 160
	p := int(pos)
	switch {
	case p < 85:
		return color.RGBA{byte(max - p*max/85), byte(p * max / 85), 0, 255}
	case p < 170:
		p -= 85
		return color.RGBA{0, byte(max - p*max/85), byte(p * max / 85), 255}
	default:
		p -= 170
		return color.RGBA{byte(p * max / 86), 0, byte(max - p*max/86), 255}
	}
}
