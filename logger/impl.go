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

import "fmt"

// Debug logs a message (with optional context) at the DEBUG log level
func Debug(msg string, ctx ...interface{}) {
	if Log != nil {
		Log.Debug(msg, ctx...)
	}
}

// Info logs a message (with optional context) at the INFO log level
func Info(msg string, ctx ...interface{}) {
	if Log != nil {
		Log.Info(msg, ctx...)
	}
}

// Warn logs a message (with optional context) at the WARNING log level
func Warn(msg string, ctx ...interface{}) {
	if Log != nil {
		Log.Warn(msg, ctx...)
	}
}

// Error logs a message (with optional context) at the ERROR log level
func Error(msg string, ctx ...interface{}) {
	if Log != nil {
		Log.Error(msg, ctx...)
	}
}

// Crit logs a message (with optional context) at the CRITICAL log level
func Crit(msg string, ctx ...interface{}) {
	if Log != nil {
		Log.Crit(msg, ctx...)
	}
}

// Infof logs at the INFO log level using a standard printf format string
func Infof(format string, args ...interface{}) {
	if Log != nil {
		Log.Info(fmt.Sprintf(format, args...))
	}
}

// Debugf logs at the DEBUG log level using a standard printf format string
func Debugf(format string, args ...interface{}) {
	if Log != nil {
		Log.Debug(fmt.Sprintf(format, args...))
	}
}

// Warnf logs at the WARNING log level using a standard printf format string
func Warnf(format string, args ...interface{}) {
	if Log != nil {
		Log.Warn(fmt.Sprintf(format, args...))
	}
}

// Errorf logs at the ERROR log level using a standard printf format string
func Errorf(format string, args ...interface{}) {
	if Log != nil {
		Log.Error(fmt.Sprintf(format, args...))
	}
}

// Critf logs at the CRITICAL log level using a standard printf format string
func Critf(format string, args ...interface{}) {
	if Log != nil {
		Log.Crit(fmt.Sprintf(format, args...))
	}
}
