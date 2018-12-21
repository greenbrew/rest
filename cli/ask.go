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
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskBool asks a question and expect a yes/no answer.
func AskBool(question string, defaultAnswer string) bool {
	for {
		answer := askQuestion(question, defaultAnswer)

		if stringInSlice(strings.ToLower(answer), []string{"yes", "y"}) {
			return true
		} else if stringInSlice(strings.ToLower(answer), []string{"no", "n"}) {
			return false
		}

		invalidInput()
	}
}

// Ask a question on the output stream and read the answer from the input stream
func askQuestion(question, defaultAnswer string) string {
	fmt.Printf(question)

	return readAnswer(defaultAnswer)
}

// Read the user's answer from the input stream, trimming newline and providing a default.
func readAnswer(defaultAnswer string) string {
	stdin := bufio.NewReader(os.Stdin)
	answer, _ := stdin.ReadString('\n')
	answer = strings.TrimSuffix(answer, "\n")
	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = defaultAnswer
	}

	return answer
}

// Print an invalid input message on the error stream
func invalidInput() {
	fmt.Fprintf(os.Stderr, "Invalid input, try again.\n\n")
}

func stringInSlice(key string, list []string) bool {
	for _, entry := range list {
		if entry == key {
			return true
		}
	}
	return false
}
