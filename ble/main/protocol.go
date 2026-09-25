package main

// Wire protocol: a single characteristic on Gopherino's custom service.
//
// Remote -> robot, written without response, always 3 bytes: cmd, a, b
//
//	cmdMode  mode, 0     switch mode (modeNormal, modeTank, modeMusic, modeAuto)
//	cmdDrive left, right wheel speeds as int8, -100..100 (Normal/Tank only)
//	cmdBeep  sound, 0    sndBeep or sndHorn (Normal/Tank only)
//	cmdFlash which, 0    flashAll, flashLeft or flashRight (Normal/Tank only)
//	cmdTone  hi, lo      buzzer frequency in Hz, uint16 big endian; 0 = silence
//	                     (Music only)
//
// Robot -> remote, notified every statusInterval, 4 bytes:
//
//	statusMsg, mode, distance in cm (255 = nothing in range), flags
//	flags bit 0 (statusObstacle): Auto mode is dodging an obstacle

import (
	"runtime/volatile"
	"time"

	"tinygo.org/x/bluetooth"
)

const (
	modeNormal = iota
	modeTank
	modeMusic
	modeAuto
	numModes
)

const (
	cmdMode  = 0x10
	cmdDrive = 0x11
	cmdBeep  = 0x12
	cmdFlash = 0x13
	cmdTone  = 0x14

	statusMsg      = 0x80
	statusObstacle = 0x01
)

const (
	sndBeep = 0
	sndHorn = 1
)

const (
	flashAll   = 0
	flashLeft  = 1
	flashRight = 2
)

var (
	gopherinoServiceUUID    = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x01, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})
	gopherinoCharUUID       = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x02, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})
	gopherinoCharacteristic bluetooth.Characteristic
)

// Commands arrive in the SoftDevice's interrupt handler, where nothing may
// allocate or block, so WriteEvent only copies them into this ring buffer
// and the main loop runs them.
type command [3]byte

var (
	cmdQueue         [32]command
	cmdHead, cmdTail volatile.Register8
)

func queueCommand(client bluetooth.Connection, offset int, value []byte) {
	if len(value) != len(command{}) {
		return
	}
	head := cmdHead.Get()
	next := (head + 1) % uint8(len(cmdQueue))
	if next == cmdTail.Get() {
		return // full, drop it
	}
	copy(cmdQueue[head][:], value)
	cmdHead.Set(next)
}

func processCommands(now time.Time) {
	for tail := cmdTail.Get(); tail != cmdHead.Get(); tail = cmdTail.Get() {
		handleCommand(cmdQueue[tail], now)
		cmdTail.Set((tail + 1) % uint8(len(cmdQueue)))
	}
}

func handleCommand(c command, now time.Time) {
	manual := mode == modeNormal || mode == modeTank
	switch c[0] {
	case cmdMode:
		if c[1] != mode {
			setMode(c[1])
		}
	case cmdDrive:
		if manual {
			setWheels(int(int8(c[1])), int(int8(c[2])))
			lastDrive = now
		}
	case cmdBeep:
		if manual {
			playSound(c[1], now)
		}
	case cmdFlash:
		if manual {
			flash(c[1], now)
		}
	case cmdTone:
		if mode == modeMusic {
			toneOn(uint16(c[1])<<8 | uint16(c[2]))
		}
	}
}

var statusBuf [4]byte

func sendStatus() {
	cm := distanceMM / 10
	if distanceMM <= 0 || cm > 254 {
		cm = 255
	}
	statusBuf[0] = statusMsg
	statusBuf[1] = mode
	statusBuf[2] = byte(cm)
	statusBuf[3] = 0
	if obstacleActive() {
		statusBuf[3] |= statusObstacle
	}
	// Errors (e.g. no remote subscribed yet) are harmless here.
	gopherinoCharacteristic.Write(statusBuf[:])
}
