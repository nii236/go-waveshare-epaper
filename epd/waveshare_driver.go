package epd

import (
	"image"
	"image/color"
	"log"
	"time"

	"github.com/MaxHalford/halfgone"
)

var cmdPanelSetting byte = 0x00
var cmdPowerSetting byte = 0x01
var cmdPowerOff byte = 0x02
var cmdPowerOffSequenceSetting byte = 0x03
var cmdPowerOn byte = 0x04
var cmdPowerOnMeasure byte = 0x05
var cmdBoosterSoftStart byte = 0x06
var cmdDeepSleep byte = 0x07
var cmdDisplayStartTransmission1 byte = 0x10
var cmdDataStop byte = 0x11
var cmdDisplayRefresh byte = 0x12
var cmdDisplayStartTransmission2 byte = 0x13
var cmdDualSPI byte = 0x15
var cmdAutoSequence byte = 0x17
var cmdKWLUToption byte = 0x2B
var cmdPLLControl byte = 0x30
var cmdTemperatureSensorCalibration byte = 0x40
var cmdTemperatureSensorSelection byte = 0x41
var cmdTemperatureSensorWrite byte = 0x42
var cmdTemperatureSensorRead byte = 0x43
var cmdPanelBreakCheck byte = 0x44
var cmdVCOMAndDataIntervalSetting byte = 0x50
var cmdLowerPowerDetection byte = 0x51
var cmdEndVoltageSetting byte = 0x52
var cmdTCONSetting byte = 0x60
var cmdResolutionSetting byte = 0x61
var cmdGateSourceStartSetting byte = 0x65
var cmdRevision byte = 0x70
var cmdGetStatus byte = 0x71
var cmdAutoMeasurementVCOM byte = 0x80
var cmdReadVCOMValue byte = 0x81
var cmdVCOMDCSetting byte = 0x82

// EPD holds driver state for e-paper module
type EPD struct {
	RPi    *RaspberryPi
	Width  int
	Height int
}

// New EPD driver
func New(
	rpi *RaspberryPi,
	Width int, // 800
	Height int, // 480

) (*EPD, error) {
	epd := &EPD{
		RPi:    rpi,
		Width:  Width,
		Height: Height,
	}
	epd.Reset()
	epd.SendCommand(cmdPowerSetting)
	epd.SendData([]byte{0x07})
	epd.SendData([]byte{0x07}) //VGH=20V,VGL=-20V
	epd.SendData([]byte{0x3f}) //VDH=15V
	epd.SendData([]byte{0x3f}) //VDL=-15V

	epd.SendCommand(cmdPowerOn)
	time.Sleep(100 * time.Millisecond)
	epd.ReadBusy()

	epd.SendCommand(cmdPanelSetting)
	epd.SendData([]byte{0x1F}) //KW-3f   KWR-2F	BWROTP 0f	BWOTP 1f

	epd.SendCommand(cmdResolutionSetting)
	epd.SendData([]byte{0x03}) //source 800
	epd.SendData([]byte{0x20})
	epd.SendData([]byte{0x01}) //gate 480
	epd.SendData([]byte{0xE0})

	epd.SendCommand(cmdDualSPI)
	epd.SendData([]byte{0x00})

	epd.SendCommand(cmdVCOMAndDataIntervalSetting) //VCOM AND DATA INTERVAL SETTING
	epd.SendData([]byte{0x10})
	epd.SendData([]byte{0x07})

	epd.SendCommand(cmdTCONSetting) //TCON SETTING
	epd.SendData([]byte{0x22})

	//EPD hardware init end
	return epd, nil
}

// Reset the EPD
func (e *EPD) Reset() {
	e.RPi.DigitalWrite(e.RPi.ResetPin, 1)
	time.Sleep(200 * time.Millisecond)
	e.RPi.DigitalWrite(e.RPi.ResetPin, 0)
	time.Sleep(2 * time.Millisecond)
	e.RPi.DigitalWrite(e.RPi.ResetPin, 1)
	time.Sleep(200 * time.Millisecond)

}

// SendCommand to the SPI Bus
func (e *EPD) SendCommand(cmd byte) {
	e.RPi.DigitalWrite(e.RPi.DcPin, 0)
	e.RPi.DigitalWrite(e.RPi.CsPin, 0)
	e.RPi.SPI.Write([]byte{cmd})
	e.RPi.DigitalWrite(e.RPi.CsPin, 1)

}

// SendData to the SPI Bus
func (e *EPD) SendData(data []byte) {
	e.RPi.DigitalWrite(e.RPi.DcPin, 1)
	e.RPi.DigitalWrite(e.RPi.CsPin, 0)
	e.RPi.SPI.Write(data)
	e.RPi.DigitalWrite(e.RPi.CsPin, 1)

}

// ReadBusy blocks until the device is no longer busy
// This kinda sucks, prefer func isBusy() bool for non-blockage
func (e *EPD) ReadBusy() {
	log.Println("e-Paper busy")
	e.SendCommand(cmdGetStatus)
	busy := MustRead(e.RPi.DigitalRead(e.RPi.BusyPin))
	for busy == 0 {
		e.SendCommand(cmdGetStatus)
		busy = MustRead(e.RPi.DigitalRead(e.RPi.BusyPin))
	}
	time.Sleep(200 * time.Millisecond)

}

// GetBuffer returns the image currently on the screen
func (e *EPD) GetBuffer() image.Image {

}

type Pixel int

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

// Display renders the image onto the screen
func (e *EPD) Display(img image.Image) {
	e.SendCommand(cmdDisplayStartTransmission2)
	result := halfgone.FloydSteinbergDitherer{}.Apply(halfgone.ImageToGray(img))

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			grayPx := color.GrayModel.Convert(img.At(x, y).RGBA())
			row = append(row, grayPx.RGBA())
		}
		pixels = append(pixels, row)
	}

	e.RPi.SPI.Write(result)
	e.SendCommand(cmdDisplayRefresh)
	time.Sleep(100 * time.Millisecond)
	e.ReadBusy()
}

// Clear the screen
func (e *EPD) Clear() {
	e.SendCommand(cmdDisplayStartTransmission1)
	for i := 0; i < e.Width*e.Height/8; i++ {
		// Send one byte at a time until it flushes all pixels
		e.SendData([]byte{0x00})
	}

	e.SendCommand(cmdDisplayStartTransmission2)
	for i := 0; i < e.Width*e.Height/8; i++ {
		// Send one byte at a time until it flushes all pixels
		e.SendData([]byte{0x00})
	}

	e.SendCommand(cmdDisplayRefresh)
	time.Sleep(100 * time.Millisecond)
	e.ReadBusy()
}

// Sleep sends EPD into deep sleep to save power
func (e *EPD) Sleep() {
	e.SendCommand(cmdPowerOff) //POWER_OFF
	e.ReadBusy()

	e.SendCommand(cmdDeepSleep) //DEEP_SLEEP
	e.SendData([]byte{0xA5})
}
