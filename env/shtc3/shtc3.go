// Driver for the SHTC3 temperature/humidity sensor
// https://www.mouser.sg/datasheet/2/682/Sensirion_Humidity_Sensors_SHTC3_Datasheet-1386761.pdf
package shtc3

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
	notConnectedErr = errors.New("sensor not connected")
	wrongDeviceID   = errors.New("wrong device ID")
	checksumErr     = errors.New("crc failed")
)

const (
	Addr          = 0x70
	dataRange     = float32(1 << 16)
	chipIDReg     = 0xEFC8
	chipIDRegMask = 0x083F
	chipIDRegV    = 0x0807
	// softResetReg  = 0x805D
	// sleepReg      = 0xB098
	// wakeUpReg     = 0x3517
)

// cmds not using stretch
const (
	// tempHumNorm = 0x7866 // normal measurement, temp first with Clock Stretch Enabled
	// tempHumLPS  = 0x609C // low power measurement, temp first with Clock Stretch Enabled
	humTempNorm = 0x58E0 // normal measurement, hum first with Clock Stretch Enabled
	// humTempLP   = 0x401A // low power measurement, hum first with Clock Stretch Enabled
)

// cmds using stretch
// const (
// 	tempHumNormStrtc = 0x7CA2 // normal measurement, temp first with Clock Stretch Enabled
// 	tempHumLPStrtc   = 0x6458 // low power measurement, temp first with Clock Stretch Enabled
// 	humTempNormStrtc = 0x5C24 // normal measurement, hum first with Clock Stretch Enabled
// 	humTempLPStrtc   = 0x44DE // low power measurement, hum first with Clock Stretch Enabled
// )

type Params struct {
	Bus i2c.Bus
}

type device struct {
	conn *i2c.Dev
	m    sync.RWMutex

	// isAsleep bool

	name string
}

func New(p Params) (d env.HumDriver, err error) {
	_, err = host.Init()
	if err != nil {
		return
	}

	d = &device{
		conn: &i2c.Dev{
			Bus:  p.Bus,
			Addr: Addr,
		},
		// set name based on i2c addr
		name: "shtc3-" + fmt.Sprintf("0x%2.2x", Addr),
	}

	// check sensor is connected
	err = d.IsConnected()

	return
}

func (d *device) String() string {
	return d.name
}

func (d *device) StringJSON() (string, error) {
	hum, temp, err := d.GetHumTemp()
	if err != nil {
		return "", err
	}
	jOpt := env.HumDevice{
		Time: time.Now().UTC().Format(time.RFC3339Nano),
		Name: d.name,
		Data: env.Hum{
			H: hum,
			T: temp},
	}
	b, err := json.Marshal(jOpt)
	return string(b), err
}

func (d *device) IsConnected() error {
	if err := d.CheckChipID(); err != nil {
		return fmt.Errorf("%v: %v", notConnectedErr, err)
	}
	return nil
}

func (d *device) CheckChipID() error {
	if v, err := readUint16FromReg(d, chipIDReg); err != nil {
		return err
	} else if v&chipIDRegMask != chipIDRegV {
		return fmt.Errorf("%v (%v)", wrongDeviceID, v)
	}
	return nil
}

func (d *device) GetHumTemp() (hum, temp float32, err error) {
	d.m.RLock()
	defer d.m.RUnlock()

	// send measurement request cmd
	err = d.conn.Tx(utils.Uint16ToBytes(humTempNorm), nil)
	if err != nil {
		return
	}

	// measurement cycle duration is ~ 12 ms
	time.Sleep(15 * time.Millisecond)

	data := make([]byte, 6)
	err = d.conn.Tx(nil, data)
	if err != nil {
		return
	}

	hum = float32(utils.BytesToUint16(data[0], data[1])) / dataRange * 100.0
	temp = float32(utils.BytesToUint16(data[3], data[4]))/dataRange*175.0 - 45.0

	if crcErr := checkCRC8(data[:3]); crcErr != nil {
		err = fmt.Errorf("%v: %v", checksumErr, crcErr)
	}
	if crcErr := checkCRC8(data[3:]); crcErr != nil {
		err = fmt.Errorf("%v: %v", checksumErr, crcErr)
	}

	return
}

func readUint16FromReg(d *device, reg uint16) (v uint16, err error) {
	d.m.RLock()
	defer d.m.RUnlock()

	reply := make([]byte, 3)
	if err = d.conn.Tx(utils.Uint16ToBytes(reg), reply); err != nil {
		return
	}

	v = utils.BytesToUint16(reply[0], reply[1])
	err = checkCRC8(reply)
	return
}

func checkCRC8(bytes []byte) error {
	l := len(bytes)
	if l != 3 {
		return fmt.Errorf("CRC needs 3 bytes")
	}

	crc := 0xFF
	for i := 0; i < l-1; i++ {
		crc ^= int(bytes[i])
		for j := 0; j < 8; j++ {
			if (crc & 0x80) != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc = crc << 1
			}
		}
	}
	crc &= 0xFF
	if crc == int(bytes[2]) {
		return nil
	} else {
		return fmt.Errorf("crc fail (0x%2.2x != 0x%2.2x)", crc, bytes[2])
	}
}

// for validation only
var _ env.HumDriver = &device{}
