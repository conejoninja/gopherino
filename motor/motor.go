package motor

import (
	"tinygo.org/x/drivers"
)

type Direction uint8
type Motor uint8

const (
	Address              = 0x10
	Forward    Direction = 0x00
	Backward   Direction = 0x01
	MotorRight Motor     = 0x02
	MotorLeft  Motor     = 0x00
)

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
	speed   uint8
}

func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		buf:     make([]byte, 5),
		Address: Address,
		speed:   50,
	}
}

func (d *Device) Configure() {
	d.buf[0] = 0x00
	d.buf[1] = uint8(Forward)
	d.buf[2] = 20
	d.buf[3] = uint8(Forward)
	d.buf[4] = 20
}

func (d *Device) SetSpeed(speed uint8) {
	d.speed = speed
}

func (d *Device) Forward() {
	d.buf[1] = uint8(Forward)
	d.buf[2] = d.speed
	d.buf[3] = uint8(Forward)
	d.buf[4] = d.speed
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) Backward() {
	d.buf[1] = uint8(Backward)
	d.buf[2] = d.speed
	d.buf[3] = uint8(Backward)
	d.buf[4] = d.speed
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) SpinRight() {
	d.buf[1] = uint8(Forward)
	d.buf[2] = d.speed
	d.buf[3] = uint8(Backward)
	d.buf[4] = d.speed
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) SpinLeft() {
	d.buf[1] = uint8(Backward)
	d.buf[2] = d.speed
	d.buf[3] = uint8(Forward)
	d.buf[4] = d.speed
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) Right() {
	d.buf[1] = uint8(Forward)
	d.buf[2] = d.speed
	d.buf[3] = uint8(Forward)
	d.buf[4] = 20
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) Left() {
	d.buf[1] = uint8(Forward)
	d.buf[2] = 20
	d.buf[3] = uint8(Forward)
	d.buf[4] = d.speed
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}

func (d *Device) Stop() {
	d.buf[1] = uint8(Forward)
	d.buf[2] = 20
	d.buf[3] = uint8(Forward)
	d.buf[4] = 20
	d.bus.Tx(uint16(d.Address), d.buf, nil)
}
