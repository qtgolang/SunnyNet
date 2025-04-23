//go:build mini
// +build mini

package Resource

import (
	_ "embed"
	"io"
)

//go:embed CertInstallDocument.html
var FrontendIndex []byte

func ReadVueFile(name string) ([]byte, error) {
	return nil, io.EOF
}
