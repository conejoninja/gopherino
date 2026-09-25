package main

// Buzzer driven by a hardware PWM, so tones play in the background while
// the main loop keeps driving and lighting: Music mode holds a note for as
// long as the remote's button is held, and beeps/horns are short sequences
// stepped by updateSound.

import (
	"machine"
	"time"
)

const (
	// The micro:bit v2's onboard speaker. Use machine.P0 for the Maqueen's
	// own buzzer instead.
	bzrPin = machine.P27

	minFreq = 100 // Hz, the lowest tone the PWM is configured for
)

var (
	bzrPWM = machine.PWM0
	bzrCh  uint8
)

type toneStep struct {
	freq uint16 // Hz, 0 = silence
	dur  time.Duration
}

var sounds = [...][]toneStep{
	sndBeep: {{1760, 80 * time.Millisecond}},
	sndHorn: {{440, 120 * time.Millisecond}, {0, 60 * time.Millisecond}, {440, 250 * time.Millisecond}},
}

var (
	seq       []toneStep
	seqIdx    int
	seqNextAt time.Time
)

func initSound() error {
	// Configure for the lowest frequency: SetPeriod can later shorten the
	// period but never go beyond this one (see machine.PWM.SetPeriod).
	if err := bzrPWM.Configure(machine.PWMConfig{Period: 1e9 / minFreq}); err != nil {
		return err
	}
	ch, err := bzrPWM.Channel(bzrPin)
	if err != nil {
		return err
	}
	bzrCh = ch
	toneOff()
	return nil
}

func toneOn(freq uint16) {
	if freq < minFreq {
		toneOff()
		return
	}
	bzrPWM.SetPeriod(1e9 / uint64(freq))
	bzrPWM.Set(bzrCh, bzrPWM.Top()/2)
}

func toneOff() {
	bzrPWM.Set(bzrCh, 0)
}

func playSound(id byte, now time.Time) {
	if int(id) >= len(sounds) {
		return
	}
	seq = sounds[id]
	seqIdx = 0
	toneOn(seq[0].freq)
	seqNextAt = now.Add(seq[0].dur)
}

func updateSound(now time.Time) {
	if seq == nil || now.Before(seqNextAt) {
		return
	}
	seqIdx++
	if seqIdx >= len(seq) {
		seq = nil
		toneOff()
		return
	}
	toneOn(seq[seqIdx].freq)
	seqNextAt = now.Add(seq[seqIdx].dur)
}

func stopSound() {
	seq = nil
	toneOff()
}
