package main

// Gopherino BLE firmware, driven by the nicectrlr "gopherino-remote" example
// (code.madriguera.me/GoEducation/nicectrlr/examples/gopherino-remote).
//
// The remote picks one of four modes:
//   - Normal: the remote mixes one joystick into two wheel speeds. LEDs blue.
//   - Tank:   each joystick drives one wheel. LEDs green.
//   - Music:  the remote's buttons play notes on the buzzer. LEDs rainbow.
//   - Auto:   obstacle avoidance with the HC-SR04. LEDs solid red while
//     driving, blinking red while dodging an obstacle.
//
// In Normal and Tank the remote's buttons beep or flash the lights for a
// moment. See protocol.go for the wire format.
//
// Flash with:
//
//	tinygo flash -target microbit-v2-s113v7 ./ble/main

import (
	"machine"
	"runtime/volatile"
	"time"

	"tinygo.org/x/bluetooth"
	"tinygo.org/x/drivers/hcsr04"

	"github.com/conejoninja/gopherino/motor"
)

const (
	tick = 20 * time.Millisecond

	// Motor speed (0..255) sent for a 100% stick deflection.
	maxSpeed = 200

	// Normal/Tank: stop the wheels if the remote goes quiet for this long
	// (out of range, crashed, ...). The remote sends a drive command at
	// least every 150ms.
	driveTimeout = 500 * time.Millisecond

	// How often the distance is measured and reported outside Auto mode.
	statusInterval = 250 * time.Millisecond
)

var (
	adapter = bluetooth.DefaultAdapter

	i2c           = machine.I2C1
	maqueenMotor  *motor.Device
	maqueenHCSR04 hcsr04.Device

	mode      byte = modeNormal
	lastDrive time.Time

	wheelL, wheelR int16 // last speeds sent to the motor driver

	distanceMM int32 // last HC-SR04 reading, 0 = nothing in range
	nextStatus time.Time

	// disconnected is set from the BLE connect handler, which runs in
	// interrupt context, and handled by the main loop.
	disconnected volatile.Register8
)

func main() {
	maqueenHCSR04 = hcsr04.New(machine.P1, machine.P2)
	maqueenHCSR04.Configure()
	i2c.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
	})
	maqueenMotor = motor.New(i2c)
	maqueenMotor.Configure()
	maqueenMotor.Stop()

	initLights()
	must("init sound", initSound())

	startBLE()

	for {
		now := time.Now()

		if disconnected.Get() != 0 {
			disconnected.Set(0)
			setMode(modeNormal)
		}
		processCommands(now)

		switch mode {
		case modeNormal, modeTank:
			if now.Sub(lastDrive) > driveTimeout {
				setWheels(0, 0)
			}
		case modeAuto:
			tickAuto(now)
		}

		updateSound(now)
		updateLights(now)
		updateStatus(now)

		time.Sleep(tick)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func startBLE() {
	adapter.SetConnectHandler(func(d bluetooth.Device, connected bool) {
		if !connected {
			disconnected.Set(1)
		}
	})

	must("enable BLE stack", adapter.Enable())
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gopherino",
	}))
	must("start adv", adv.Start())

	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: gopherinoServiceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &gopherinoCharacteristic,
				UUID:   gopherinoCharUUID,
				Flags: bluetooth.CharacteristicReadPermission |
					bluetooth.CharacteristicWritePermission |
					bluetooth.CharacteristicWriteWithoutResponsePermission |
					bluetooth.CharacteristicNotifyPermission,
				WriteEvent: queueCommand,
			},
		},
	}))
}

// setMode switches to m, always starting from a stopped, silent robot.
func setMode(m byte) {
	if m >= numModes {
		return
	}
	mode = m
	setWheels(0, 0)
	stopSound()
	resetAuto()
	resetFlashes()
	println("mode", m)
}

// setWheels drives each wheel at a signed percentage (-100..100) of
// maxSpeed. The motor driver is only written to when something changed.
func setWheels(left, right int) {
	l := int16(clamp(left, -100, 100) * maxSpeed / 100)
	r := int16(clamp(right, -100, 100) * maxSpeed / 100)
	if l == wheelL && r == wheelR {
		return
	}
	wheelL, wheelR = l, r
	maqueenMotor.Drive(l, r)
}

// updateStatus measures the distance (Auto mode already does it every tick)
// and notifies it to the remote, every statusInterval.
func updateStatus(now time.Time) {
	if now.Before(nextStatus) {
		return
	}
	nextStatus = now.Add(statusInterval)
	if mode != modeAuto {
		measureDistance()
	}
	sendStatus()
}

func measureDistance() int32 {
	distanceMM = maqueenHCSR04.ReadDistance()
	return distanceMM
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
