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
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// FileResponseEntry is a file transfer response
type FileResponseEntry struct {
	Identifier string
	Path       string
	Filename   string
	Buffer     []byte /* either a path or a buffer must be provided */
}

type fileResponse struct {
	req              *http.Request
	files            []FileResponseEntry
	headers          map[string]string
	removeAfterServe bool
}

func (r *fileResponse) Render(w http.ResponseWriter) error {
	if r.headers != nil {
		for k, v := range r.headers {
			w.Header().Set(k, v)
		}
	}

	switch len(r.files) {
	case 0:
		return nil
	case 1:
		return r.renderInline(w)
	default:
		return r.renderMultipart(w)
	}
}

func (r *fileResponse) String() string {
	return fmt.Sprintf("%d files", len(r.files))
}

// FileResponse returns a response renderer for a file download
func FileResponse(r *http.Request, files []FileResponseEntry, headers map[string]string, removeAfterServe bool) Response {
	return &fileResponse{r, files, headers, removeAfterServe}
}

func (r *fileResponse) renderInline(w http.ResponseWriter) error {
	var rs io.ReadSeeker
	var mt time.Time
	var sz int64

	if r.files[0].Path == "" {
		rs = bytes.NewReader(r.files[0].Buffer)
		mt = time.Now()
		sz = int64(len(r.files[0].Buffer))
	} else {
		f, err := os.Open(r.files[0].Path)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		mt = fi.ModTime()
		sz = fi.Size()
		rs = f
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", sz))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline;filename=%s", r.files[0].Filename))

	http.ServeContent(w, r.req, r.files[0].Filename, mt, rs)
	if r.files[0].Path != "" && r.removeAfterServe {
		err := os.Remove(r.files[0].Path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *fileResponse) renderMultipart(w http.ResponseWriter) error {
	// Now the complex multipart answer
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	for _, entry := range r.files {
		var rd io.Reader
		if entry.Path != "" {
			fd, err := os.Open(entry.Path)
			if err != nil {
				return err
			}
			defer fd.Close()
			rd = fd
		} else {
			rd = bytes.NewReader(entry.Buffer)
		}

		fw, err := mw.CreateFormFile(entry.Identifier, entry.Filename)
		if err != nil {
			return err
		}

		_, err = io.Copy(fw, rd)
		if err != nil {
			return err
		}
	}
	mw.Close()

	w.Header().Set("Content-Type", mw.FormDataContentType())
	w.Header().Set("Content-Length", fmt.Sprintf("%d", body.Len()))

	_, err := io.Copy(w, body)
	return err
}
