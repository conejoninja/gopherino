package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hcsr04"

	"github.com/conejoninja/gopherino/motor"
)

var (
	d                   int32
	maqueen_hcsr04      hcsr04.Device
	i2c                 = machine.I2C0
	maqueen_motor       *motor.Device
	lineLeft, lineRight machine.Pin
)

func main() {
	maqueen_hcsr04 = hcsr04.New(machine.P1, machine.P2)
	maqueen_hcsr04.Configure()
	i2c.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_100KHZ})
	maqueen_motor = motor.New(i2c)
	maqueen_motor.Configure()

	// CONFIGURE PINS for the infrared grayscale sensor (line followers)
	/*lineLeft = machine.P13
	lineRight = machine.P14
	lineLeft.Configure(machine.PinConfig{Mode: machine.PinInput})
	lineRight.Configure(machine.PinConfig{Mode: machine.PinInput})*/

	d = 123
	for true {
		d = maqueen_hcsr04.ReadDistance()

		if d < 60 {
			maqueen_motor.Stop()
			maqueen_motor.SpinRight()
		} else {
			maqueen_motor.Stop()
			maqueen_motor.Forward()
		}

		time.Sleep(500 * time.Millisecond)
	}
}
