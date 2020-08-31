/*
	Copyright (c) 2020 platy11.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, version 3.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"
)

// MagnetometerAxis is the axis (0-2: [x, y, z]) that is used to determine
// whether the doog door is open
const MagnetometerAxis = 1

// DoorClosedFieldStrength is the expected value for the specified axis when
// the dog door is closed
const DoorClosedFieldStrength = -270

// DoorClosedFieldRange is the range of values around DoorClosedFieldStrength
// where the dog door will be considered to be closed
const DoorClosedFieldRange = 70

// CheckInterval is the interval between each poll of the sensor, specified in
// milliseconds
const CheckInterval = 3500

// TorchFeatEnabled specifies whether the torch feature will be used if the
// door is opened
const TorchFeatEnabled = true

// Timezone is the IANA Time Zone database name for the local time zone
const Timezone = "Australia/Melbourne"

// TorchHourMax is the maximum hour when the torch will be used.
// 0 <= TorchHourMax <= 23
const TorchHourMax = 6

// TorchHourMin is the minimum hour when the torch will be used.
// 0 <= TorchHourMin <= 23
const TorchHourMin = 18

// DogNamePossessive is the dog's name that will be used in messages
const DogNamePossessive = "Puppy's"

// TgBotToken is the Bot API Token assigned by Telegram
var TgBotToken = os.Getenv("DETECTOR_TG_KEY")

// TgChatID is the ID of the Telegram chat (usually a group) that the program
// will send messages to. Note that this will be negative for groups, and will
// change if the group is 'upgraded to a supergroup'
var TgChatID = os.Getenv("DETECTOR_TG_CHAT")

// MessagePrefix will be prepended to all messages sent. This is useful for
// testing the program.
var MessagePrefix = os.Getenv("DETECTOR_MSG_PREFIX") + " "

// DoorUsedMessage is the mesage that will be sent when the dog door is used
var DoorUsedMessage = MessagePrefix + DogNamePossessive +
	" dog door was used\\!"

// DoorBlockedMessage is the mesage that will be sent when the dog door has
// been open for multiple subsequent periods
var DoorBlockedMessage = MessagePrefix + DogNamePossessive +
	" dog door has been open for over 45 seconds\\. " +
	"It may be blocked or the detector may have malfunctioned\\."

// openCount is the number of subsequent periods when the door has been open
var openCount uint = 0

// lightCount is the number of subsequent periods when the light has been on.
// -1 indicates that the light is currently disabled.
var lightCount int = -1

var hasReceivedData = false
var location *time.Location

// MagnetometerData holds the data received from the magnetic sensor for each
// axis, i.e. [x, y, z]
type MagnetometerData [3]float64

type magnetometerDataRaw struct {
	Data struct {
		Values [3]float64 `json:"Values"`
	} `json:"AK8963 Magnetometer"`
}

func main() {
	log.Println("Dog Door Detector version 0.1 - starting")
	log.Println("Press Ctrl+D to exit")

	var err error // Prevent shadowing global-scoped `location`
	location, err = time.LoadLocation(Timezone)
	if err != nil {
		panic(err)
	}

	log.Println("Waiting 5s to allow time for positioning...")
	time.Sleep(5 * time.Second)
	log.Println("Starting")

	outputJSONBlocks := make(chan []byte, 2)
	stopSensor := make(chan bool, 1)

	go gracefulEnd(stopSensor)

	go StreamSensorValues(CheckInterval, outputJSONBlocks, stopSensor)
	for block := range outputJSONBlocks {
		var rawdata magnetometerDataRaw
		err := json.Unmarshal(block, &rawdata)
		if err != nil {
			log.Fatal(fmt.Errorf("json.Unmarshal JSON block: %w", err))
		}

		var data MagnetometerData = rawdata.Data.Values
		open := DoorOpen(data)
		DoorTick(data)
		if open {
			OpenDoorTick(data)
		} else {
			ClosedDoorTick(data)
		}
	}

	// Make sure the torch is off before the process ends
	if err := SetTorch(false); err != nil {
		log.Fatal(err)
	}
	if err := Cleanup(); err != nil {
		log.Fatal(err)
	}
	log.Println("Process completed")
}

func gracefulEnd(stopSensor chan<- bool) {
	for {
		var discard string
		_, err := fmt.Scan(&discard)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(fmt.Errorf("fmt.Scan stdin: %w", err))
		}
		time.Sleep(2 * time.Second)
	}
	stopSensor <- true
	log.Printf("EOF received, process ending")
}

// DoorOpen returns a bool of the state of the dog door based on the configured
// calibration variables and supplied sensor readings
func DoorOpen(data MagnetometerData) bool {
	v := data[MagnetometerAxis]
	return !(math.Abs(v-float64(DoorClosedFieldStrength)) < DoorClosedFieldRange)
}

// UseTorch returns whether the torch should be used at this time. l should
// be the local time zone.
func UseTorch(l *time.Location) bool {
	if !TorchFeatEnabled {
		return false
	}
	h := time.Now().In(l).Hour()
	if TorchHourMax > TorchHourMin {
		return h < TorchHourMax && h >= TorchHourMin
	}
	return h < TorchHourMax || h >= TorchHourMin
}

// DoorTick is run each time the sensor is polled
func DoorTick(_ MagnetometerData) {
	if !hasReceivedData {
		log.Println("Ready")
		hasReceivedData = true
	}
	if lightCount >= 0 {
		lightCount++
	}
	if lightCount >= 7 {
		// The door has been open for 7 * CheckInterval, disable the light
		lightCount = -1
		go func() {
			err := SetTorch(false)
			if err != nil {
				log.Println(fmt.Errorf("turn off torch: %w", err))
			}
		}()
	}
}

// ClosedDoorTick is run each time the sensor is polled and the door is shut
func ClosedDoorTick(data MagnetometerData) {
	openCount = 0
	// fmt.Printf("CLOSED: [-, %v] %v\n", LightCount, data)
	if lightCount >= 4 {
		lightCount = -1
		go func() {
			err := SetTorch(false)
			if err != nil {
				log.Println(fmt.Errorf("turn off torch: %w", err))
			}
		}()
	}
}

// OpenDoorTick is run each time the sensor is polled and the door is open
func OpenDoorTick(data MagnetometerData) {
	openCount++
	log.Printf("  OPEN: [%v, %v] %v\n", openCount, lightCount, data)

	if openCount == 1 {
		go func() {
			err := TgSendMessage(TgBotToken, TgChatID, DoorUsedMessage, false)
			if err != nil {
				log.Println(fmt.Errorf("send telegram message: %w", err))
			}
		}()
	}

	if openCount == 14 {
		// The door has been open for 14 * CheckInterval, it's probably blocked
		go func() {
			err := TgSendMessage(TgBotToken, TgChatID, DoorBlockedMessage, false)
			if err != nil {
				log.Println(fmt.Errorf("send telegram message: %w", err))
			}
		}()
	}

	if UseTorch(location) && openCount == 1 {
		lightCount = 0
		go func() {
			err := SetTorch(true)
			if err != nil {
				log.Println(fmt.Errorf("turn on torch: %w", err))
			}
		}()
	}
}
