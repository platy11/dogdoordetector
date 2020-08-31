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
	"fmt"
	"os/exec"
)

// SetTorch sets the torch to the specified state
func SetTorch(state bool) error {
	var arg string
	if state {
		arg = "on"
	} else {
		arg = "off"
	}
	cmd := exec.Command("termux-torch", arg)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exec `termux-torch`: %w", err)
	}
	return nil
}
