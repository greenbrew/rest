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

package api

import (
	"time"
)

// Operation represents a background operation as result of a REST POST, PATCH,
// DELETE or PUT
type Operation struct {
	ID          string                 `json:"id" yaml:"id"`
	Description string                 `json:"description" yaml:"description"`
	CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
	Status      string                 `json:"status" yaml:"status"`
	StatusCode  StatusCode             `json:"status_code" yaml:"status_code"`
	Resources   map[string][]string    `json:"resources" yaml:"resources"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
	Err         string                 `json:"err" yaml:"err"`
}
