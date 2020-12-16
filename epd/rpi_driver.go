package epd

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

// RaspberryPi holds the embed GPIO state
type RaspberryPi struct {
	ResetPin int
	DcPin    int
	BusyPin  int
	CsPin    int
	Host     *embd.Descriptor
	SPI      embd.SPIBus
}

// New RPI driver
func (p *RaspberryPi) New(
	ResetPin int,
	DcPin int,
	BusyPin int,
	CsPin int,
	SPIChannel byte, //0
	SPISpeed int, //4000000
	SPIBPW int, //8
	SPIDelay int, //0
) (*RaspberryPi, error) {
	embd.InitGPIO()
	embd.InitSPI()
	host, err := embd.DescribeHost()
	if err != nil {
		return nil, err
	}
	SPIBus := embd.NewSPIBus(
		embd.SPIMode0,
		SPIChannel,
		SPISpeed,
		SPIBPW,
		SPIDelay,
	)

	pi := &RaspberryPi{
		ResetPin: ResetPin,
		DcPin:    DcPin,
		BusyPin:  BusyPin,
		CsPin:    CsPin,
		Host:     host,
		SPI:      SPIBus,
	}
	var pin embd.DigitalPin
	pin, err = host.GPIODriver().DigitalPin(ResetPin)
	pin.SetDirection(embd.Out)
	pin, err = host.GPIODriver().DigitalPin(DcPin)
	pin.SetDirection(embd.Out)
	pin, err = host.GPIODriver().DigitalPin(BusyPin)
	pin.SetDirection(embd.In)
	pin, err = host.GPIODriver().DigitalPin(CsPin)
	pin.SetDirection(embd.Out)
	if err != nil {
		return nil, err
	}
	return pi, nil

}

// DigitalWrite to an output
func (p *RaspberryPi) DigitalWrite(pinNum int, value int) error {
	pin, err := p.Host.GPIODriver().DigitalPin(pinNum)
	if err != nil {
		return err
	}
	err = pin.Write(value)
	if err != nil {
		return err
	}
	return nil
}

// DigitalRead from an input
func (p *RaspberryPi) DigitalRead(pinNum int) (int, error) {
	pin, err := p.Host.GPIODriver().DigitalPin(pinNum)
	if err != nil {
		return 0, err
	}
	return pin.Read()
}

// SPIWritebyte writes data to serial
func (p *RaspberryPi) SPIWritebyte(data []byte) error {
	_, err := p.SPI.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Close cleans up the connections
func (p *RaspberryPi) Close() error {
	var err error
	err = p.DigitalWrite(p.ResetPin, 0)
	if err != nil {
		return err
	}
	err = p.DigitalWrite(p.DcPin, 0)
	if err != nil {
		return err
	}
	err = p.Host.GPIODriver().Close()
	if err != nil {
		return err
	}
	err = p.SPI.Close()
	if err != nil {
		return err
	}
	err = p.Host.SPIDriver().Close()
	if err != nil {
		return err
	}
	return nil
}
