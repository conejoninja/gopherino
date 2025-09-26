package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/bluetooth"
	"tinygo.org/x/drivers/uc8151"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/notosans"
)

var (
	gopherinoServiceUUID = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x01, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})
	gopherinoCharUUID    = bluetooth.NewUUID([16]byte{0xa0, 0xb4, 0x00, 0x02, 0x92, 0x6d, 0x4d, 0x61, 0x98, 0xdf, 0x8c, 0x5c, 0x62, 0xee, 0x53, 0x72})

	display          uc8151.Device
	black            = color.RGBA{1, 1, 1, 255}
	white            = color.RGBA{0, 0, 0, 255}
	lastButtonStates = [5]bool{false, false, false, false, false}
	buttonPins       = [5]machine.Pin{machine.BUTTON_A, machine.BUTTON_B, machine.BUTTON_C, machine.BUTTON_UP, machine.BUTTON_DOWN}

	adapter         = bluetooth.DefaultAdapter
	connectedDevice bluetooth.Device
	gopherinoChar   bluetooth.DeviceCharacteristic
)

const WIDTH = 296
const HEIGHT = 128

func main() {

	led3v3 := machine.ENABLE_3V3
	led3v3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led3v3.High()

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 12000000,
		SCK:       machine.EPD_SCK_PIN,
		SDO:       machine.EPD_SDO_PIN,
	})

	display = uc8151.New(machine.SPI0, machine.EPD_CS_PIN, machine.EPD_DC_PIN, machine.EPD_RESET_PIN, machine.EPD_BUSY_PIN)
	display.Configure(uc8151.Config{
		Rotation: uc8151.ROTATION_270,
		Speed:    uc8151.MEDIUM,
		Blocking: false,
	})

	display.ClearBuffer()
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, 10, 16, "NOT CONNECTED", black)
	_, w := tinyfont.LineWidth(&notosans.Notosans12pt, "FORWARD")
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, WIDTH-int16(w)-6, 36, "FORWARD", black)
	_, w = tinyfont.LineWidth(&notosans.Notosans12pt, "BACKWARD")
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, WIDTH-int16(w)-6, 100, "BACKWARD", black)
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, 20, 124, "LEFT", black)
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, WIDTH-50, 124, "RIGHT", black)
	_, w = tinyfont.LineWidth(&notosans.Notosans12pt, "PARTY")
	tinyfont.WriteLine(&display, &notosans.Notosans12pt, (WIDTH-int16(w))/2, 124, "PARTY", black)
	display.Display()
	display.WaitUntilIdle()

	println("Starting Badger2040W BLE as gamepad...")
	setupButtons()

	must("habilitar BLE", adapter.Enable())
	println("🔍 Searching for 'Gopherino C3:C3:6C:81:24:11'...")

	ch := make(chan bluetooth.ScanResult, 1)

	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == "C3:C3:6C:81:24:11" {
			println("✅ FoundGopherino:", result.Address.String())

			adapter.StopScan()
			ch <- result
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			println(err.Error())
			return
		}
		display.FillRectangle(0, 0, 200, 20, white)
		tinyfont.WriteLine(&display, &notosans.Notosans12pt, 10, 16, "CONNECTED", black)
		display.Display()
		display.WaitUntilIdle()

		println("connected to ", result.Address.String())
	}

	// get services
	println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{gopherinoServiceUUID})
	must("discover services", err)

	if len(srvcs) == 0 {
		panic("could not find heart rate service")
	}

	srvc := srvcs[0]

	println("found service", srvc.UUID().String())

	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{gopherinoCharUUID})
	if err != nil {
		println(err)
	}

	if len(chars) == 0 {
		panic("could not find heart rate characteristic")
	}

	gopherinoChar = chars[0]
	println("found characteristic", gopherinoChar.UUID().String())

	for {
		handleButtons()
		time.Sleep(50 * time.Millisecond)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

func setupButtons() {
	buttonPins[0].Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	buttonPins[1].Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	buttonPins[2].Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	buttonPins[3].Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	buttonPins[4].Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
}

func handleButtons() {
	currentButtonStates := [5]bool{}
	for i := range buttonPins {
		currentButtonStates[i] = buttonPins[i].Get()
	}

	for i := 0; i < 5; i++ {

		if currentButtonStates[i] != lastButtonStates[i] {
			if currentButtonStates[i] {
				sendButton(0x01, byte(i+1))
			} else {
				sendButton(0x02, byte(i+1))
			}
		}
	}

	lastButtonStates = currentButtonStates
}

func sendButton(action, buttonID byte) {
	data := []byte{action, buttonID}

	n, err := gopherinoChar.WriteWithoutResponse(data)
	if err != nil {
		println("Error sending press event:", err.Error())
	} else {
		println("Press event sent", buttonID, action, n)
	}
}
