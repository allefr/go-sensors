// Driver for the MCP9808 temperature sensor
// https://www.microchip.com/wwwproducts/en/en556182
package mcp9808

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/allefr/go-sensors/env"
	"github.com/allefr/go-sensors/utils"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/host"
)

var (
	wrongI2cAddr    = errors.New("i2c address must be within 0x18 and 0x1F")
	notConnectedErr = errors.New("sensor not connected")
	wrongManufID    = errors.New("wrong manufacturer ID")
	wrongDeviceID   = errors.New("wrong device ID")
)

const (
	ambTempReg = 0x05
	manIdReg   = 0x06
	devIdReg   = 0x07
)

const (
	manIdRegV = 0x0054
	devIdRegV = 0x0400
)

type Params struct {
	Bus  i2c.Bus
	Addr uint16
	Name string
}

type device struct {
	conn *i2c.Dev
	m    sync.RWMutex

	name string
}

func New(p Params) (d env.TempDriver, err error) {
	_, err = host.Init()
	if err != nil {
		return
	}

	// sanity check on provided address
	if p.Addr < 0x18 || p.Addr > 0x1F {
		return nil, wrongI2cAddr
	}

	// if name not provided, make one using i2c addr
	n := "mcp9808-" + fmt.Sprintf("0x%2.2x", p.Addr)
	if p.Name != "" {
		n = p.Name
	}

	d = &device{
		conn: &i2c.Dev{
			Bus:  p.Bus,
			Addr: p.Addr,
		},
		name: n,
	}

	// check sensor is connected
	err = d.IsConnected()

	return
}

func (d *device) String() string {
	return d.name
}

func (d *device) StringJSON() (string, error) {
	temp, err := d.GetTemp()
	if err != nil {
		return "", err
	}
	jOpt := env.TempDevice{
		Time: time.Now().UTC().Format(time.RFC3339Nano),
		Name: d.name,
		Data: env.Temp{T: temp},
	}
	b, err := json.Marshal(jOpt)
	return string(b), err
}

func (d *device) IsConnected() error {
	if err := d.CheckManufID(); err != nil {
		return fmt.Errorf("%v: %v", notConnectedErr, err)
	}
	if err := d.CheckDeviceID(); err != nil {
		return fmt.Errorf("%v: %v", notConnectedErr, err)
	}
	return nil
}

func (d *device) CheckManufID() error {
	if v, err := readUint16(d, manIdReg); err != nil {
		return err
	} else if v != manIdRegV {
		return fmt.Errorf("%v (%v != %v)", wrongManufID, v, manIdRegV)
	}
	return nil
}

func (d *device) CheckDeviceID() error {
	if v, err := readUint16(d, devIdReg); err != nil {
		return err
	} else if v&devIdRegV != devIdRegV {
		// checking first byte only, cause 2nd is revision, which could change
		return fmt.Errorf("%v (%v != %v)", wrongDeviceID, v, devIdRegV)
	}
	return nil
}

func (d *device) GetTemp() (temp float32, err error) {
	v, err := readUint16(d, ambTempReg)
	if err != nil {
		return
	}
	temp = float32(v&0x0FFF) / 16.0
	if (v & 0x1000) > 0 {
		temp -= 256.0
	}

	return
}

func readUint16(d *device, reg int) (v uint16, err error) {
	d.m.RLock()
	defer d.m.RUnlock()

	reply := make([]byte, 2)
	if err = d.conn.Tx([]byte{byte(reg)}, reply); err != nil {
		return
	}

	v = utils.BytesToUint16(reply[0], reply[1])
	return
}

// for validation only
var _ env.TempDriver = &device{}
