// +build tiny

package uiserver

import (
	"fmt"
	"net/http"
)

func assetFS() http.FileSystem {
	return noFS{}
}

type noFS struct{}

func (noFS) Open(_ string) (http.File, error) {
	return nil, fmt.Errorf("no file")
}
