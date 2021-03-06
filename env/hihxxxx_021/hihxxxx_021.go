// Driver for the HIH6030 temperature/humidity sensor
// https://sensing.honeywell.com/HIH6030-021-001-humidity-sensors
// https://sensing.honeywell.com/i2c-comms-humidicon-tn-009061-2-en-final-07jun12.pdf
package hihxxxx_021

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
	warnStaleData   = errors.New("reading stale data")
	errCommandMode  = errors.New("device in command mode")
)

const (
	Addr      = 0x27 // can ask manufacturer to custom, otherwise this is it
	dataRange = float32(1<<14) - 2
	readReg   = 0x00
)

const (
	isStaleData = 1 << (iota + 6)
	isCommandMode
)

type Params struct {
	Bus  i2c.Bus
	Name string
}

type device struct {
	conn *i2c.Dev
	m    sync.RWMutex

	name string
}

func New(p Params) (d env.HumDriver, err error) {
	_, err = host.Init()
	if err != nil {
		return
	}

	// if name not provided, make one using i2c addr
	n := "hih0000-021-" + fmt.Sprintf("0x%2.2x", Addr)
	if p.Name != "" {
		n = p.Name
	}

	d = &device{
		conn: &i2c.Dev{
			Bus:  p.Bus,
			Addr: Addr,
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
	if _, _, err := d.GetHumTemp(); err != nil {
		return notConnectedErr
	}
	return nil
}

func (d *device) GetHumTemp() (hum, temp float32, err error) {
	d.m.RLock()
	defer d.m.RUnlock()

	// send measurement request cmd
	err = d.conn.Tx([]byte{readReg}, nil)
	if err != nil {
		return
	}

	// measurement cycle duration is ~ 36.65 ms
	time.Sleep(40 * time.Millisecond)

	data := make([]byte, 4)
	err = d.conn.Tx(nil, data)
	if err != nil {
		return
	}

	hum = float32(utils.BytesToUint16(data[0], data[1])&0x3FFF) / dataRange * 100.0
	temp = float32(utils.BytesToUint16(data[2], data[3])>>2)/dataRange*165.0 - 40.0

	// check status bits
	if (data[0] & isCommandMode) == isCommandMode {
		err = errCommandMode
	}
	if (data[0] & isStaleData) == isStaleData {
		err = warnStaleData
	}

	return
}

// for validation only
var _ env.HumDriver = &device{}
