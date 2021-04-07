package env

import (
	"encoding/json"
)

type HumDriver interface {
	Driver
	GetHumTemp() (float32, float32, error)
}

type Hum struct {
	T float32 `json:"temp"`
	H float32 `json:"hum"`
}

func (g *Hum) String() (string, error) {
	b, err := json.Marshal(g)
	return string(b), err
}

type HumDevice struct {
	Time string `json:"datetime"`
	Name string `json:"sensor"`
	Data Hum    `json:"data"`
}

func (g *HumDevice) String() (string, error) {
	b, err := json.Marshal(g)
	return string(b), err
}
