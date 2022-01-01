package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	VendorID  = 0x04d9
	ProductID = 0xa052
	co2op     = 0x50
	tempop    = 0x42
)

type Data struct {
	Time int64    `json:"time"`
	Co2  *int     `json:"co2"`
	Temp *float64 `json:"temp"`
}

var (
	key = []byte{0x86, 0x41, 0xc9, 0xa8, 0x7f, 0x41, 0x3c, 0xac}
	interval time.Duration
)

func init() {
	flag.DurationVar(&interval, "i", time.Duration(5*time.Second), "interval")
}

func validate(buf []byte) bool {
	if len(buf) < 5 {
		return false
	}

	if buf[4] != 0x0d {
		return false
	}
	if (buf[0]+buf[1]+buf[2])&0xff != buf[3] {
		return false
	}
	return true
}

func monitor(w io.Writer) error {
	if err := hid.Init(); err != nil {
		return fmt.Errorf("failed to hid.Init: %w", err)
	}
	defer hid.Exit()

	device, err := hid.OpenFirst(VendorID, ProductID)
	if err != nil {
		return fmt.Errorf("failed to hid.OpenFirst: %w", err)
	}
	defer device.Close()

	if _, err := device.SendFeatureReport(key); err != nil {
		return fmt.Errorf("failed to device.SendFeatureReport: %w", err)
	}

	ch := make(chan Data)
	go func() {
		buf := make([]byte, 8)
		for {
			_, err := device.Read(buf)
			if err != nil {
				log.Println("failed to read:", err)
				continue
			}

			dec := decrypt(buf, key)
			if !validate(dec) {
				if !validate(buf) {
					log.Printf("failed to decrypt: %x", buf)
					continue
				}
				dec = buf
			}
			val := int(dec[1])<<8 | int(dec[2])

			var data Data
			switch dec[0] {
			case co2op:
				data.Co2 = &val
			case tempop:
				temp := float64(val)/16.0 - 273.15
				data.Temp = &temp
			default:
				continue
			}
			ch <- data
		}
	}()

	enc := json.NewEncoder(w)
	tick := time.Tick(interval)
	var data Data
	for {
		select {
		case d, ok := <-ch:
			if !ok {
				return nil
			}
			if d.Co2 != nil {
				data.Co2 = d.Co2
			}
			if d.Temp != nil {
				data.Temp = d.Temp
			}
		case <-tick:
			data.Time = time.Now().Unix()
			if err := enc.Encode(data); err != nil {
				return fmt.Errorf("failed to enc.Encode: %w", err)
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()
	log.Fatal(monitor(os.Stdout))
}
