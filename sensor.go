/*
	Copyright (c) 2020 platy11.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, version 3.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
	GNU General Public License for more details	.

	You should have received a copy of the GNU General Public License
	along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// StreamSensorValues runs `termux-sensor` and sends each JSON output as a byte
// slice representing text to the provided channel
func StreamSensorValues(interval uint, output chan<- []byte, stop <-chan bool) {
	intervalArg := strconv.FormatUint(uint64(interval), 10)

	cmd := exec.Command("termux-sensor", "-s", "magnet", "-d", intervalArg)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(fmt.Errorf("cmd.StdoutPipe termux-sensor: %w", err))
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(fmt.Errorf("cmd.Start termux-sensor: %w", err))
	}

	scanner := bufio.NewScanner(stdout)
	buffer := make([][]byte, 0, 9)

	for scanner.Scan() {
		if stopIfSignal(stop, cmd, output) {
			break
		}
		byteData := scanner.Bytes()
		buffer = append(buffer, byteData)
		if len(buffer) != 9 {
			continue
		}
		output <- bytes.Join(buffer, nil)
		buffer = nil
	}
}

func stopIfSignal(stop <-chan bool, cmd *exec.Cmd, output chan<- []byte) bool {
	select {
	case _ = <-stop:
		// Sending os.Interrupt and waiting doesn't seem to work, potentially
		// related to the other signal issues?
		// Waiting for termux-sensor to exit neatly would remove the need to
		// do sensor cleanup later
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal(fmt.Errorf("cmd.Process.Signal to termux-sensor: %w", err))
		}
		// Close the channel, allowing the main goroutine to end
		close(output)
		return true
	default:
		return false
	}
}

// Cleanup runs termux-sensor -c to clean up/detach sensors after use
// This is used at the end of the program because we can not cleanly end
// termux-sensor; see above
func Cleanup() error {
	cmd := exec.Command("termux-sensor", "-c")
	log.Printf("Cleaning up sensors...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run termux-sensor -c: %w", err)
	}
	return nil
}
