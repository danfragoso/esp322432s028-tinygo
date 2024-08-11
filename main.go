package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/xpt2046"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freesans"
)

var (
	LCD_SCK_PIN = machine.GPIO14
	LCD_SDO_PIN = machine.GPIO13
	LCD_SDI_PIN = machine.GPIO12
	LCD_DC_PIN  = machine.GPIO2
	LCD_SS_PIN  = machine.GPIO15
	LCD_RST_PIN = machine.NoPin
)

func InitDisplay() drivers.Displayer {
	machine.SPI2.Configure(machine.SPIConfig{
		SCK:       LCD_SCK_PIN,
		SDO:       LCD_SDO_PIN,
		SDI:       LCD_SDI_PIN,
		Frequency: 40000000,
	})

	d := ili9341.NewSPI(
		machine.SPI2,
		LCD_DC_PIN,
		LCD_SS_PIN,
		LCD_RST_PIN,
	)
	d.Configure(ili9341.Config{
		Width:            320,
		Height:           240,
		Rotation:         drivers.Rotation180,
		DisplayInversion: false,
	})
	d.FillScreen(color.RGBA{0, 120, 233, 255})

	machine.GPIO21.Configure(machine.PinConfig{machine.PinOutput})
	machine.GPIO21.High()

	return d
}

var (
	TOUCH_CLK_PIN  = machine.GPIO25 // CLK
	TOUCH_CS_PIN   = machine.GPIO33 // CS
	TOUCH_DIN_PIN  = machine.GPIO32 // DIN
	TOUCH_DOUT_PIN = machine.GPIO39 // DOUT
	TOUCH_IRQ_PIN  = machine.GPIO36 // IRQ
)

func InitTouch() xpt2046.Device {
	t := xpt2046.New(TOUCH_CLK_PIN, TOUCH_CS_PIN, TOUCH_DIN_PIN, TOUCH_DOUT_PIN, TOUCH_IRQ_PIN)
	t.Configure(&xpt2046.Config{
		Precision: 10, //Maximum number of samples for a single ReadTouchPoint to improve accuracy.
	})

	return t
}

type touchPoint struct {
	X int
	Y int
}

func main() {
	display := InitDisplay()
	touch := InitTouch()

	colors := []color.RGBA{
		{255, 255, 255, 255},
		{255, 255, 0, 255},
		{0, 255, 0, 255},
		{0, 255, 255, 255},
		{0, 0, 255, 255},
		{255, 0, 255, 255},
		{255, 255, 255, 255},
	}

	tinyfont.WriteLineColors(display, &freesans.Regular18pt7b, 24, 36, "Hello from TinyGO", colors)
	tinyfont.WriteLine(display, &freesans.Regular18pt7b, 24, 92, "touch me", colors[0])

	// create channel to receive touch events
	touchEvents := make(chan touchPoint)

	// start a goroutine to read touch events
	go func() {
		for {
			for !touch.Touched() {
				time.Sleep(50 * time.Millisecond)
			}

			point := touch.ReadTouchPoint()
			touchEvents <- touchPoint{X: 320 - ((point.Y * 320) >> 16), Y: (point.X * 240) >> 16}

			for touch.Touched() {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case point := <-touchEvents:
			if mapPointToButton(point) == "touch me" {
				tinyfont.WriteLine(display, &freesans.Regular18pt7b, 24, 120, "touched", colors[0])
			}

		default:
			display.Display()
			time.Sleep(33 * time.Millisecond)
		}
	}
}

func mapPointToButton(point touchPoint) string {
	if point.Y > 72 && point.Y < 100 {
		return "touch me"
	}

	return "none"
}
