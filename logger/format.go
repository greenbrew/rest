// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Roberto Mier Escandon <rmescandon@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// Pretty will attempt to convert any Go structure into a string suitable for logging
func Pretty(input interface{}) string {
	pretty, err := json.MarshalIndent(input, "\t", "\t")
	if err != nil {
		return fmt.Sprintf("%v", input)
	}

	return fmt.Sprintf("\n\t%s", pretty)
}

// GetStack will convert the Go stack into a string suitable for logging
func GetStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)

	return fmt.Sprintf("\n\t%s", buf[:n])
}
