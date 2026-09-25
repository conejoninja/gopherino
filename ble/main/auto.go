package main

// Auto mode: drive forward until the HC-SR04 sees something close, then
// back up and spin away from it before trying again.

import "time"

const (
	obstacleDistance = 150 // mm
	autoSpeed        = 40  // % of maxSpeed
	backupTime       = 400 * time.Millisecond
	turnTime         = 600 * time.Millisecond
)

type autoState uint8

const (
	autoCruise autoState = iota
	autoBackup
	autoTurn
)

var (
	autoSt    autoState
	autoUntil time.Time
)

func resetAuto() {
	autoSt = autoCruise
}

// obstacleActive reports whether Auto mode is busy dodging an obstacle.
func obstacleActive() bool {
	return mode == modeAuto && autoSt != autoCruise
}

func tickAuto(now time.Time) {
	switch autoSt {
	case autoCruise:
		// 0 means the echo timed out: nothing in range.
		if d := measureDistance(); d > 0 && d < obstacleDistance {
			autoSt = autoBackup
			autoUntil = now.Add(backupTime)
			setWheels(-autoSpeed, -autoSpeed)
		} else {
			setWheels(autoSpeed, autoSpeed)
		}
	case autoBackup:
		if now.After(autoUntil) {
			autoSt = autoTurn
			autoUntil = now.Add(turnTime)
			// Pick a side pseudo-randomly so it doesn't get stuck in a corner.
			if now.UnixNano()&1 == 0 {
				setWheels(autoSpeed, -autoSpeed)
			} else {
				setWheels(-autoSpeed, autoSpeed)
			}
		}
	case autoTurn:
		if now.After(autoUntil) {
			autoSt = autoCruise
		}
	}
}
