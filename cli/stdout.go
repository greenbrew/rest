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

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

// PrintResources prints out a list of resources, in a well formatted list
// one per row and padded
func PrintResources(resources map[string][]string) {
	for k, r := range resources {
		fmt.Printf("%s:\n", k)
		for _, j := range r {
			fmt.Printf("  - %s\n", path.Base(j))
		}
	}
}

// ConfirmDeletion prints out a question about removing resources and
// returns error if 'yes' is not replied
func ConfirmDeletion(what string, deletableItems []string) error {
	fmt.Printf("The following %v will be REMOVED:\n", what)
	for _, it := range deletableItems {
		fmt.Printf("  - %s\n", it)
	}

	if !AskBool("Do you want to continue? [Y/n]: ", "yes") {
		return fmt.Errorf("Abort")
	}
	return nil
}

// DumpData prints out data in yaml or json format, as desired
func DumpData(format string, data interface{}) error {
	if format == "yaml" {
		b, err := yaml.Marshal(&data)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, string(b))
	} else if format == "json" {
		b, err := json.Marshal(&data)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, string(b))
	} else {
		return fmt.Errorf("Unsupported format '%s'", format)
	}
	return nil
}
