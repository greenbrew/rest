package api

import "path/filepath"

// Version of the builtin API
const Version = "1.0"

// Path prefixes API version to a path
func Path(path ...string) string {
	p := append([]string{"/", Version}, path...)
	return filepath.Join(p...)
}
