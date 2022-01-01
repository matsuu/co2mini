package main

import (
	"encoding/json"
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

var (
	key = []byte{0x86, 0x41, 0xc9, 0xa8, 0x7f, 0x41, 0x3c, 0xac}
)

type Data struct {
	Time int64    `json:"time"`
	Co2  *int     `json:"co2"`
	Temp *float64 `json:"temp"`
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

	enc := json.NewEncoder(w)
	buf := make([]byte, 8)

	for {
		_, err := device.Read(buf)
		if err != nil {
			log.Println("failed to read:", err)
			continue
		}
		now := time.Now().Unix()

		dec := decrypt(buf, key)
		// TODO: checksum
		if dec[4] != 0x0d {
			if buf[4] != 0x0d {
				log.Println("failed to decrypt:", buf)
				continue
			}
			dec = buf
		}
		val := int(dec[1])<<8 | int(dec[2])

		data := Data{Time: now}
		switch dec[0] {
		case co2op:
			data.Co2 = &val
		case tempop:
			temp := float64(val)/16.0 - 273.15
			data.Temp = &temp
		default:
			continue
		}
		if err := enc.Encode(data); err != nil {
			return fmt.Errorf("failed to enc.Encode: %w", err)
		}
	}
	return nil
}

func main() {
	log.Fatal(monitor(os.Stdout))
}
