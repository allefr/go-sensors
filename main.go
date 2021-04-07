package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allefr/go-sensors/env/hihxxxx_021"
	"github.com/allefr/go-sensors/env/mcp9808"
	"github.com/allefr/go-sensors/env/shtc3"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
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

	// mcp9808
	p := mcp9808.Params{Bus: bus, Addr: 0x1a}
	tS, err := mcp9808.New(p)
	if err != nil {
		log.Fatalf("%s: %v\n", tS.String(), err)
	}

	// hih6030
	hS, err := hihxxxx_021.New(hihxxxx_021.Params{Bus: bus})
	if err != nil {
		log.Fatalf("%s: %v\n", hS.String(), err)
	}

	// shtc3
	hS2, err := shtc3.New(shtc3.Params{Bus: bus})
	if err != nil {
		log.Fatalf("%s: %v\n", hS2.String(), err)
	}

	// start infinite loop to query data
	for {
		if str, err := tS.StringJSON(); err != nil {
			fmt.Printf("%s: %v\n", tS.String(), err)
		} else {
			fmt.Println(str)
		}

		if str, err := hS.StringJSON(); err != nil {
			fmt.Printf("%s: %v\n", hS.String(), err)
		} else {
			fmt.Println(str)
		}

		if str, err := hS2.StringJSON(); err != nil {
			fmt.Printf("%s: %v\n", hS2.String(), err)
		} else {
			fmt.Println(str)
		}

		// wait some time
		time.Sleep(1 * time.Second)
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
