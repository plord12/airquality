package main

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "sen5x_i2c.h"
// #include "sensirion_common.h"
// #include "sensirion_i2c_hal.h"
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"time"
	"unsafe"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jessevdk/go-flags"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type Options struct {
	Broker    string `short:"b" long:"broker" description:"MQTT broker address" required:"true"`
	Username  string `short:"u" long:"username" description:"MQTT broker username" required:"true"`
	Password  string `short:"p" long:"password" description:"MQTT broker password" required:"true"`
	BaseTopic string `short:"t" long:"basetopic" description:"MQTT base topic" default:"homeassistant/sensor/airquality"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func main() {
	// parse flags
	//
	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(options.Broker).SetClientID("airquality")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts = opts.SetUsername(options.Username)
	opts = opts.SetPassword(options.Password)
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// create MQTT configs
	config, _ := json.Marshal(map[string]string{
		"name":                "Air Quality PM1.0",
		"unique_id":           "air_quality_pm1p0",
		"object_id":           "air_quality_pm1p0",
		"unit_of_measurement": "µg/m³",
		"state_topic":         options.BaseTopic + "/air_quality_pm1p0/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_pm1p0/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":                "Air Quality PM2.5",
		"unique_id":           "air_quality_pm2p5",
		"object_id":           "air_quality_pm2p5",
		"unit_of_measurement": "µg/m³",
		"state_topic":         options.BaseTopic + "/air_quality_pm2p5/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_pm2p5/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":                "Air Quality PM4.0",
		"unique_id":           "air_quality_pm4p0",
		"object_id":           "air_quality_pm4p0",
		"unit_of_measurement": "µg/m³",
		"state_topic":         options.BaseTopic + "/air_quality_pm4p0/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_pm4p0/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":                "Air Quality PM10",
		"unique_id":           "air_quality_pm10p0",
		"object_id":           "air_quality_pm10p0",
		"unit_of_measurement": "µg/m³",
		"state_topic":         options.BaseTopic + "/air_quality_pm10p0/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_pm10p0/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":                "Air Quality Humidity",
		"unique_id":           "air_quality_humidity",
		"object_id":           "air_quality_humidity",
		"unit_of_measurement": "%",
		"state_topic":         options.BaseTopic + "/air_quality_humidity/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_humidity/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":                "Air Quality Temperature",
		"unique_id":           "air_quality_temperature",
		"object_id":           "air_quality_temperature",
		"unit_of_measurement": "°C",
		"state_topic":         options.BaseTopic + "/air_quality_temperature/state",
		"value_template":      "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_temperature/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":           "Air Quality VOC",
		"unique_id":      "air_quality_voc",
		"object_id":      "air_quality_voc",
		"state_topic":    options.BaseTopic + "/air_quality_voc/state",
		"value_template": "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_voc/config", 0, true, config).Wait()
	config, _ = json.Marshal(map[string]string{
		"name":           "Air Quality NOX",
		"unique_id":      "air_quality_nox",
		"object_id":      "air_quality_nox",
		"state_topic":    options.BaseTopic + "/air_quality_nox/state",
		"value_template": "{{ value }}",
	})
	c.Publish(options.BaseTopic+"/air_quality_nox/config", 0, true, config).Wait()

	C.sensirion_i2c_hal_init(C.CString("/dev/i2c-0"))

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

	var xaxis []time.Time
	var pm1p0 []float64
	var pm2p5 []float64
	var pm4p0 []float64
	var pm10p0 []float64
	var humidity []float64
	var temperature []float64
	var voc []float64
	var nox []float64
	lastHour := 0

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
		c.Publish(options.BaseTopic+"/air_quality_pm1p0/state", 0, false, fmt.Sprintf("%.1f", mass_concentration_pm1p0)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_pm2p5/state", 0, false, fmt.Sprintf("%.1f", mass_concentration_pm2p5)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_pm4p0/state", 0, false, fmt.Sprintf("%.1f", mass_concentration_pm4p0)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_pm10p0/state", 0, false, fmt.Sprintf("%.1f", mass_concentration_pm10p0)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_humidity/state", 0, false, fmt.Sprintf("%.1f", ambient_humidity)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_temperature/state", 0, false, fmt.Sprintf("%.1f", ambient_temperature)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_voc/state", 0, false, fmt.Sprintf("%.1f", voc_index)).Wait()
		c.Publish(options.BaseTopic+"/air_quality_nox/state", 0, false, fmt.Sprintf("%.1f", nox_index)).Wait()

		today := time.Now()
		if today.Hour() < lastHour {
			xaxis = xaxis[:0]
			pm1p0 = pm1p0[:0]
			pm2p5 = pm2p5[:0]
			pm4p0 = pm4p0[:0]
			pm10p0 = pm10p0[:0]
			humidity = humidity[:0]
			temperature = temperature[:0]
			voc = voc[:0]
			nox = nox[:0]
		}
		lastHour = today.Hour()
		xaxis = append(xaxis, today)
		pm1p0 = append(pm1p0, float64(mass_concentration_pm1p0))
		pm2p5 = append(pm2p5, float64(mass_concentration_pm2p5))
		pm4p0 = append(pm4p0, float64(mass_concentration_pm4p0))
		pm10p0 = append(pm10p0, float64(mass_concentration_pm10p0))
		humidity = append(humidity, float64(ambient_humidity))
		temperature = append(temperature, float64(ambient_temperature))
		voc = append(voc, float64(voc_index))
		if math.IsNaN(float64(nox_index)) {
			nox = append(nox, 0)
		} else {
			nox = append(nox, float64(nox_index))
		}

		// charts
		//

		// PM 1.0
		//
		graph := chart.Chart{
			Title:      "PM1.0",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "µg/m³",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: pm1p0,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 10 {
							return drawing.Color{R: 55, G: 172, B: 86, A: 255}
						}
						if y < 20 {
							return drawing.Color{R: 155, G: 212, B: 68, A: 255}
						}
						if y < 25 {
							return drawing.Color{R: 241, G: 210, B: 8, A: 255}
						}
						if y < 50 {
							return drawing.Color{R: 255, G: 187, B: 1, A: 255}
						}
						if y < 75 {
							return drawing.Color{R: 255, G: 140, B: 0, A: 255}
						}
						return drawing.Color{R: 237, G: 15, B: 5, A: 255}
					}},
				},
			},
		}
		f, _ := os.CreateTemp("", "pm1p0*.png")
		os.Chmod(f.Name(), 0666)
		error := graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd := exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/pm1p0-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/pm1p0-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/pm1p0-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// PM 2.5
		//
		graph = chart.Chart{
			Title:      "PM2.5",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "µg/m³",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: pm2p5,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 10 {
							return drawing.Color{R: 55, G: 172, B: 86, A: 255}
						}
						if y < 20 {
							return drawing.Color{R: 155, G: 212, B: 68, A: 255}
						}
						if y < 25 {
							return drawing.Color{R: 241, G: 210, B: 8, A: 255}
						}
						if y < 50 {
							return drawing.Color{R: 255, G: 187, B: 1, A: 255}
						}
						if y < 75 {
							return drawing.Color{R: 255, G: 140, B: 0, A: 255}
						}
						return drawing.Color{R: 237, G: 15, B: 5, A: 255}
					}},
				},
			},
		}
		f, _ = os.CreateTemp("", "pm2p5*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/pm2p5-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/pm2p5-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/pm2p5-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// PM 4.0
		//
		graph = chart.Chart{
			Title:      "PM4.0",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "µg/m³",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: pm4p0,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 20 {
							return drawing.Color{R: 55, G: 172, B: 86, A: 255}
						}
						if y < 40 {
							return drawing.Color{R: 155, G: 212, B: 68, A: 255}
						}
						if y < 50 {
							return drawing.Color{R: 241, G: 210, B: 8, A: 255}
						}
						if y < 100 {
							return drawing.Color{R: 255, G: 187, B: 1, A: 255}
						}
						if y < 150 {
							return drawing.Color{R: 255, G: 140, B: 0, A: 255}
						}
						return drawing.Color{R: 237, G: 15, B: 5, A: 255}
					}},
				},
			},
		}
		f, _ = os.CreateTemp("", "pm4p0*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/pm4p0-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/pm4p0-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/pm4p0-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// PM 10.0
		//
		graph = chart.Chart{
			Title:      "PM10.0",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "µg/m³",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: pm10p0,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 20 {
							return drawing.Color{R: 55, G: 172, B: 86, A: 255}
						}
						if y < 40 {
							return drawing.Color{R: 155, G: 212, B: 68, A: 255}
						}
						if y < 50 {
							return drawing.Color{R: 241, G: 210, B: 8, A: 255}
						}
						if y < 100 {
							return drawing.Color{R: 255, G: 187, B: 1, A: 255}
						}
						if y < 150 {
							return drawing.Color{R: 255, G: 140, B: 0, A: 255}
						}
						return drawing.Color{R: 237, G: 15, B: 5, A: 255}
					}},
				},
			},
		}
		f, _ = os.CreateTemp("", "pm10p0*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/pm10p0-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/pm10p0-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/pm10p0-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// Humidity
		//
		graph = chart.Chart{
			Title:      "Humidity",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "%",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: humidity,
					Style:   chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColor: chart.ColorBlue},
				},
			},
		}
		f, _ = os.CreateTemp("", "humidity*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/humidity-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/humidity-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/humidity-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// Temperature
		//
		graph = chart.Chart{
			Title:      "Temperature",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "°C",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: temperature,
					Style:   chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColor: chart.ColorBlue},
				},
			},
		}
		f, _ = os.CreateTemp("", "temperature*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/temperature-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/temperature-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/temperature-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// VOC
		//
		graph = chart.Chart{
			Title:      "VOC",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "index",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: voc,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 249 {
							return chart.ColorBlue
						}
						if y < 449 {
							return chart.ColorGreen
						}
						return chart.ColorRed
					}},
				},
			},
		}
		f, _ = os.CreateTemp("", "voc*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/voc-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/voc-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/voc-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// NOX
		//
		graph = chart.Chart{
			Title:      "NOX",
			Background: chart.Style{Padding: chart.Box{Top: 20, Left: 20, Right: 20, Bottom: 20}},
			XAxis: chart.XAxis{
				Style: chart.Style{TextRotationDegrees: 90.0, FontSize: 6},
				ValueFormatter: func(v interface{}) string {
					typed := v.(float64)
					typedDate := chart.TimeFromFloat64(typed)
					return typedDate.Format("Jan-02-06 15:04")
				},
			},
			YAxis: chart.YAxis{
				Name:      "index",
				NameStyle: chart.Style{FontColor: chart.ColorBlack},
			},
			Series: []chart.Series{
				chart.TimeSeries{
					YAxis:   chart.YAxisPrimary,
					XValues: xaxis,
					YValues: nox,
					Style: chart.Style{StrokeColor: chart.ColorBlack, DotWidth: 3, DotColorProvider: func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
						if y < 249 {
							return chart.ColorBlue
						}
						if y < 449 {
							return chart.ColorGreen
						}
						return chart.ColorRed
					}},
				},
			},
		}
		f, _ = os.CreateTemp("", "nox*.png")
		os.Chmod(f.Name(), 0666)
		error = graph.Render(chart.PNG, f)
		if error != nil {
			fmt.Printf("render error=%v\n", error)
		}
		f.Close()
		cmd = exec.Command("scp", "-p", f.Name(), "arm3:/var/www/html/airquality/nox-"+today.Format("2006-01-02")+".png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("scp error=%v\n", error)
		}
		cmd = exec.Command("ssh", "arm3", "ln", "-sf", "/var/www/html/airquality/nox-"+today.Format("2006-01-02")+".png", "/var/www/html/airquality/nox-today.png")
		error = cmd.Run()
		if error != nil {
			fmt.Printf("ln error=%v\n", error)
		}
		os.Remove(f.Name())

		// FIX THIS ... need to run this periodically, not sleep
		//
		time.Sleep(1 * time.Minute)
	}

}
