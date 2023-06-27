package main

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "sen5x_i2c.h"
// #include "sensirion_common.h"
// #include "sensirion_i2c_hal.h"
import "C"
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"unsafe"
)

func main() {

	// Flags
	var (
		homeAssistantURL = flag.String("ha-url", "http://homeassistant.plord.co.uk:8123/api/webhook/airquality", "URL of home assistant api")
	)
	flag.Usage = func() {
		fmt.Printf("Read SEN5x air quality and send to home assistant\n\nUsage: %s [flags]\n\nWhere [flags] can be:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(*homeAssistantURL) == 0 {
		log.Println("Home assistant URL must be provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	C.sensirion_i2c_hal_init()

	error := C.sen5x_device_reset()
	if error == -1 {
		fmt.Printf("sen5x_device_reset error=%v\n", error)
	}

	serial_number := C.malloc(C.sizeof_char * 32)
	defer C.free(unsafe.Pointer(serial_number))
	serial_number_size := C.uchar(32)
	error = C.sen5x_get_serial_number((*C.uchar)(serial_number), serial_number_size)
	if error == -1 {
		fmt.Printf("sen5x_get_serial_number error=%v\n", error)
	} else {
		b := C.GoBytes(serial_number, (C.int)(serial_number_size))
		fmt.Printf("Serial number %s\n", b)
	}

	product_name := C.malloc(C.sizeof_char * 32)
	defer C.free(unsafe.Pointer(product_name))
	product_name_size := C.uchar(32)
	error = C.sen5x_get_product_name((*C.uchar)(product_name), product_name_size)
	if error == -1 {
		fmt.Printf("sen5x_get_product_name error=%v\n", error)
	} else {
		b := C.GoBytes(product_name, (C.int)(product_name_size))
		fmt.Printf("Product name %s\n", b)
	}

	error = C.sen5x_start_measurement()
	if error == -1 {
		fmt.Printf("sen5x_start_measurement error=%v\n", error)
	} else {
		fmt.Printf("Measuments started\n")
	}

	mass_concentration_pm1p0 := C.float(0.0)
	mass_concentration_pm2p5 := C.float(0.0)
	mass_concentration_pm4p0 := C.float(0.0)
	mass_concentration_pm10p0 := C.float(0.0)
	ambient_humidity := C.float(0.0)
	ambient_temperature := C.float(0.0)
	voc_index := C.float(0.0)
	nox_index := C.float(0.0)

	time.Sleep(1 * time.Second)

	// loop forever

	for {
		error = C.sen5x_read_measured_values(
			&mass_concentration_pm1p0, &mass_concentration_pm2p5,
			&mass_concentration_pm4p0, &mass_concentration_pm10p0,
			&ambient_humidity, &ambient_temperature, &voc_index, &nox_index)
		if error == -1 {
			fmt.Printf("sen5x_read_measured_values error=%v\n", error)
		} else {
			fmt.Printf("Mass concentration pm1p0: %.1f µg/m³\n", mass_concentration_pm1p0)
			fmt.Printf("Mass concentration pm2p5: %.1f µg/m³\n", mass_concentration_pm2p5)
			fmt.Printf("Mass concentration pm4p0: %.1f µg/m³\n", mass_concentration_pm4p0)
			fmt.Printf("Mass concentration pm10p0: %.1f µg/m³\n", mass_concentration_pm10p0)
			fmt.Printf("Ambient humidity: %.1f %%RH\n", ambient_humidity)
			fmt.Printf("Ambient temperature: %.1f °C\n", ambient_temperature)
			fmt.Printf("Voc index: %.1f\n", voc_index)
			fmt.Printf("Nox index: %.1f\n", nox_index)
		}

		// call home assistant
		postBody, _ := json.Marshal(map[string]string{
			"mass_concentration_pm1p0":  fmt.Sprintf("%.1f", mass_concentration_pm1p0),
			"mass_concentration_pm2p5":  fmt.Sprintf("%.1f", mass_concentration_pm2p5),
			"mass_concentration_pm4p0":  fmt.Sprintf("%.1f", mass_concentration_pm4p0),
			"mass_concentration_pm10p0": fmt.Sprintf("%.1f", mass_concentration_pm10p0),
			"ambient_humidity":          fmt.Sprintf("%.1f", ambient_humidity),
			"ambient_temperature":       fmt.Sprintf("%.1f", ambient_temperature),
			"voc_index":                 fmt.Sprintf("%.1f", voc_index),
			"nox_index":                 fmt.Sprintf("%.1f", nox_index),
		})
		responseBody := bytes.NewBuffer(postBody)
		_, error := http.Post(*homeAssistantURL, "application/json", responseBody)
		if error != nil {
			fmt.Printf("post error=%v\n", error)
		}

		// FIX THIS - create graphs for web

		time.Sleep(1 * time.Minute)
	}

}
