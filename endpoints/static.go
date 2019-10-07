// +build dev

package endpoints

import "net/http"

// Assets contains project assets.
var Assets http.FileSystem = http.Dir("static")
