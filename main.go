package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allefr/go-sensors/env"
	"github.com/allefr/go-sensors/env/hihxxxx_021"
	"github.com/allefr/go-sensors/env/mcp9808"
	"github.com/allefr/go-sensors/env/shtc3"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

const interval = 5 // [sec]

const (
	thermalTestDevice  = false
	thermalTestChamber = true
)

func main() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// open i2c bus
	bus, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		log.Fatalf("failed to open I2C: %v", err)
	}
	defer bus.Close()
	setupCloseHandler(bus)

	var t1, t2, t3, t4 env.TempDriver
	var h1, h2 env.HumDriver

	if thermalTestDevice {
		// mcp9808
		t1, err = mcp9808.New(mcp9808.Params{Bus: bus, Addr: 0x18, Name: "bread-board"})
		if err != nil {
			log.Fatalf("%s: %v\n", t1.String(), err)
		}
		t2, err = mcp9808.New(mcp9808.Params{Bus: bus, Addr: 0x19, Name: "back-bone-plate"})
		if err != nil {
			log.Fatalf("%s: %v\n", t2.String(), err)
		}
		t3, err = mcp9808.New(mcp9808.Params{Bus: bus, Addr: 0x1b, Name: "beacon"})
		if err != nil {
			log.Fatalf("%s: %v\n", t3.String(), err)
		}
		t4, err = mcp9808.New(mcp9808.Params{Bus: bus, Addr: 0x1e, Name: "I/O-chamber"})
		if err != nil {
			log.Fatalf("%s: %v\n", t4.String(), err)
		}

		// hih6030
		h1, err = hihxxxx_021.New(hihxxxx_021.Params{Bus: bus, Name: "EDFA-back"})
		if err != nil {
			log.Fatalf("%s: %v\n", h1.String(), err)
		}
	}

	if thermalTestChamber {
		// shtc3
		h2, err = shtc3.New(shtc3.Params{Bus: bus, Name: "chamber"})
		if err != nil {
			log.Fatalf("%s: %v\n", h2.String(), err)
		}
	}

	var tStart time.Time

	// start infinite loop to query data
	for {
		tStart = time.Now()

		if thermalTestDevice {
			if str, err := t1.StringJSON(); err != nil {
				// fmt.Printf("%s: %v\n", t1.String(), err)
			} else {
				fmt.Println(str)
			}
			if str, err := t2.StringJSON(); err != nil {
				// fmt.Printf("%s: %v\n", t2.String(), err)
			} else {
				fmt.Println(str)
			}
			if str, err := t3.StringJSON(); err != nil {
				// fmt.Printf("%s: %v\n", t3.String(), err)
			} else {
				fmt.Println(str)
			}
			if str, err := t4.StringJSON(); err != nil {
				// fmt.Printf("%s: %v\n", t4.String(), err)
			} else {
				fmt.Println(str)
			}

			if str, err := h1.StringJSON(); err != nil {
				// fmt.Printf("%s: %v\n", h1.String(), err)
			} else {
				fmt.Println(str)
			}
		}

		if thermalTestChamber {
			if str, err := h2.StringJSON(); err != nil {
				fmt.Printf("%s: %v\n", h2.String(), err)
			} else {
				fmt.Println(str)
			}
		}

		// repeat exactly every interval sec
		time.Sleep(interval*time.Second - time.Since(tStart))
	}

}

// make sure i2c bus gets closed even if system stops application
func setupCloseHandler(bus i2c.BusCloser) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(b i2c.BusCloser) {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		b.Close()
		os.Exit(0)
	}(bus)
}
