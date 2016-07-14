package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/i2c"
	"github.com/hybridgroup/gobot/platforms/intel-iot/edison"
)

var button *gpio.GroveButtonDriver
var buzzer *gpio.GroveBuzzerDriver
var touch *gpio.GroveTouchDriver
var rotary *gpio.GroveRotaryDriver
var sensor *gpio.GroveTemperatureSensorDriver
var sound *gpio.GroveSoundSensorDriver
var light *gpio.GroveLightSensorDriver
var lcd *i2c.GroveLcdDriver

var commands []string
var passing bool
var state int
var gameOn bool
var twist int
var heat int
var timer int

func main() {
	gbot := gobot.NewGobot()

	a := api.NewAPI(gbot)
	a.Start()

	// digital
	board := edison.NewEdisonAdaptor("edison")

	button = gpio.NewGroveButtonDriver(board, "button", "2")
	buzzer = gpio.NewGroveBuzzerDriver(board, "buzzer", "7")
	touch = gpio.NewGroveTouchDriver(board, "touch", "8")

	// analog
	rotary = gpio.NewGroveRotaryDriver(board, "rotary", "0")
	light = gpio.NewGroveLightSensorDriver(board, "light", "1")
	sound = gpio.NewGroveSoundSensorDriver(board, "sound", "2")
	sensor = gpio.NewGroveTemperatureSensorDriver(board, "sensor", "3")

	// lcd
	lcd = i2c.NewGroveLcdDriver(board, "lcd")
	commands = []string{"Switch It!", "Cover It!", "Twist It!", "Press It!", "Scream It!"}

	work := func() {

		lcd.SetRGB(255, 255, 0)
		lcd.Write("Welcome to Gleb It!")
		passing = true
		timer = 2000
		gameOn = true

		gobot.On(button.Event(gpio.Push), func(data interface{}) {
			if !gameOn {
				lcd.Clear()
				lcd.Write("Welcome to")
				lcd.SetPosition(16)
				lcd.Write("Gleb It!")
				passing = true
				timer = 2000
				gameOn = true
				state = -1
			}
			if state == 0 {
				fmt.Println("Switched!")
				passing = true
			}
		})

		gobot.On(touch.Event(gpio.Push), func(data interface{}) {
			if state == 3 {
				fmt.Println("Pressed!")
				passing = true
			}
		})

		gobot.On(light.Event("data"), func(data interface{}) {
			if state == 1 && data.(int) <= 60 {
				fmt.Println("Covered!")
				passing = true
			}
		})

		gobot.On(sound.Event("data"), func(data interface{}) {
			if state == 4 && data.(int) >= 500 {
				fmt.Println("Screamed!")
				passing = true
			}
		})

		gobot.Every(time.Duration(timer)*time.Millisecond, func() {
			if gameOn {
				if state == 2 {
					newTwist, _ := rotary.Read()
					passing = math.Abs(float64(newTwist-twist)) > 50
				}
				fmt.Println(passing)
				if !passing {
					buzzer.Tone(gpio.F4, 1)
					gameOn = false
					lcd.Clear()
					lcd.Write("Sorry! Try again!")
					return
				}
				passing = false
				lcd.Clear()
				buzzer.Tone(gpio.F4, 0.01)
				num := rand.Intn(len(commands))
				// num := 4
				state = num
				lcd.Write(commands[num])
				if state == 2 {
					twist, _ = rotary.Read()
				}
			} else {
				lcd.Clear()
				lcd.Write("Welcome to")
				lcd.SetPosition(16)
				lcd.Write("Gleb It!")
			}
		})
	}

	robot := gobot.NewRobot("glebit",
		[]gobot.Connection{board},
		[]gobot.Device{button, buzzer, touch, rotary, sensor, sound, light, lcd},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
