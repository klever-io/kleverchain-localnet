package keys

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

var pemHeaderRe = regexp.MustCompile(`-----BEGIN PRIVATE KEY for (.*?)-----`)

func ParsePEMHeader(data []byte) (string, error) {
	m := pemHeaderRe.FindSubmatch(data)
	if len(m) != 2 {
		return "", fmt.Errorf("%w: PEM does not contain expected BEGIN PRIVATE KEY header", errs.ErrInvalidInput)
	}
	return string(bytes.TrimSpace(m[1])), nil
}
