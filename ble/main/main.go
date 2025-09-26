package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/bluetooth"
	"tinygo.org/x/drivers/buzzer"
	"tinygo.org/x/drivers/hcsr04"
	"tinygo.org/x/drivers/ws2812"

	"github.com/conejoninja/gopherino/motor"
)

const (
	STOP = iota
	LEFT
	PARTY
	RIGHT
	FORWARD
	BACKWARD
	AUTOMODE
)

var (
	adapter                 = bluetooth.DefaultAdapter
	adv                     *bluetooth.Advertisement
	gopherinoServiceUUID    = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x01, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})
	gopherinoCharUUID       = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x02, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})
	gopherinoCharacteristic bluetooth.Characteristic

	d                   int32
	maqueenHCSR04       hcsr04.Device
	i2c                 = machine.I2C1
	maqueenMotor        *motor.Device
	lineLeft, lineRight machine.Pin
	gopherinoStatus     byte = STOP
	isPartying          bool

	neo    machine.Pin = machine.P15
	leds   [4]color.RGBA
	rg     bool
	ws     ws2812.Device
	bzrPin machine.Pin = machine.P27
	bzr    buzzer.Device
	song   = []note{
		{buzzer.E4, buzzer.Eighth},
		{buzzer.E4, buzzer.Eighth},
		{buzzer.Rest, buzzer.Eighth},
		{buzzer.E4, buzzer.Eighth},

		{buzzer.Rest, buzzer.Eighth},
		{buzzer.C4, buzzer.Eighth},
		{buzzer.E4, buzzer.Quarter},
		{buzzer.G4, buzzer.Quarter},

		{buzzer.Rest, buzzer.Eighth},
		{buzzer.G3, buzzer.Quarter},
	}
)

type note struct {
	tone     float64
	duration float64
}

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

	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws = ws2812.NewWS2812(neo)

	bzrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	bzr = buzzer.New(bzrPin)

	startBLE()

	for {
		switch gopherinoStatus {
		case STOP:
			maqueenMotor.Stop()
			break
		case FORWARD:
			maqueenMotor.Stop()
			maqueenMotor.Forward()
			break
		case BACKWARD:
			maqueenMotor.Stop()
			maqueenMotor.Backward()
			break
		case LEFT:
			maqueenMotor.Stop()
			maqueenMotor.SpinLeft()
			break
		case RIGHT:
			maqueenMotor.Stop()
			maqueenMotor.SpinRight()
			break
		case PARTY:
			maqueenMotor.Stop()
			partyMode()
			break
		case AUTOMODE:
			d = maqueenHCSR04.ReadDistance()
			if d < 60 {
				maqueenMotor.Stop()
				maqueenMotor.SpinRight()
			} else {
				maqueenMotor.Stop()
				maqueenMotor.Forward()
			}
			break
		}
		time.Sleep(500 * time.Millisecond)

	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func startBLE() {
	println("starting")
	time.Sleep(200 * time.Millisecond)
	must("enable BLE stack", adapter.Enable())
	adv = adapter.DefaultAdvertisement()
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
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if len(value) == 2 {
						if value[0] == 1 {
							gopherinoStatus = value[1]
							println(value[1], PARTY, isPartying)
							if value[1] == PARTY && isPartying {
								gopherinoStatus = AUTOMODE
							}
						}
						if value[0] == 2 && value[1] != PARTY {
							gopherinoStatus = STOP
						}
					}
				},
			},
		},
	}))

}

func partyMode() {
	isPartying = true
	go func() {
		for isPartying {
			rg = !rg
			for i := range leds {
				rg = !rg
				if rg {
					// Alpha channel is not supported by WS2812 so we leave it out
					leds[i] = color.RGBA{R: 0xff, G: 0x00, B: 0x00}
				} else {
					leds[i] = color.RGBA{R: 0x00, G: 0xff, B: 0x00}
				}
			}

			ws.WriteColors(leds[:])
			time.Sleep(100 * time.Millisecond)
		}
		for i := range leds {
			leds[i] = color.RGBA{R: 0x00f, G: 0x00, B: 0xff}
		}
		ws.WriteColors(leds[:])
	}()

	for _, val := range song {
		bzr.Tone(val.tone, val.duration/2)
		time.Sleep(10 * time.Millisecond)
	}
	println(gopherinoStatus, AUTOMODE, PARTY, isPartying)

	if gopherinoStatus != AUTOMODE {
		gopherinoStatus = STOP
	}
	isPartying = false
}
