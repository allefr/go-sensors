package env

import (
	"encoding/json"
)

type TempDriver interface {
	Driver
	GetTemp() (float32, error)
}

type Temp struct {
	T float32 `json:"temp"`
}

func (g *Temp) String() (string, error) {
	b, err := json.Marshal(g)
	return string(b), err
}

type TempDevice struct {
	Time string `json:"datetime"`
	Name string `json:"sensor"`
	Data Temp   `json:"data"`
}

func (g *TempDevice) String() (string, error) {
	b, err := json.Marshal(g)
	return string(b), err
}
