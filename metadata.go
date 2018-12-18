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

package rest

import (
	"fmt"
	"reflect"
)

func parseMetadata(metadata interface{}) (map[string]interface{}, error) {
	newMetadata := make(map[string]interface{})
	s := reflect.ValueOf(metadata)
	if !s.IsValid() {
		return nil, nil
	}

	if s.Kind() == reflect.Map {
		for _, k := range s.MapKeys() {
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("Invalid metadata provided (key isn't a string)")
			}
			newMetadata[k.String()] = s.MapIndex(k).Interface()
		}
	} else if s.Kind() == reflect.Ptr && !s.Elem().IsValid() {
		return nil, nil
	} else {
		return nil, fmt.Errorf("Invalid metadata provided (type isn't a map)")
	}

	return newMetadata, nil
}
