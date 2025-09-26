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
	i2c                 = machine.I2C1
	maqueen_motor       *motor.Device
	lineLeft, lineRight machine.Pin
)

func main() {
	time.Sleep(2 * time.Second)

	maqueen_hcsr04 = hcsr04.New(machine.P1, machine.P2)
	maqueen_hcsr04.Configure()
	i2c.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SCL:       machine.SCL1_PIN,
		SDA:       machine.SDA1_PIN,
	})
	maqueen_motor = motor.New(i2c)
	maqueen_motor.Configure()

	// CONFIGURE PINS for the infrared grayscale sensor (line followers)
	/*lineLeft = machine.P13
	lineRight = machine.P14
	lineLeft.Configure(machine.PinConfig{Mode: machine.PinInput})
	lineRight.Configure(machine.PinConfig{Mode: machine.PinInput})*/
	println("HOLI")
	d = 123
	for true {
		d = maqueen_hcsr04.ReadDistance()
		println("DISTANCE", d)
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
