package keys_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/keys"
)

func TestParsePEMHeader_Validator(t *testing.T) {
	data := []byte(`-----BEGIN PRIVATE KEY for abcd1234bls-----
MIGkAgEAMBAGByqGSM49AgEGBSuBBAAKBG0wawIBAQQg...
-----END PRIVATE KEY for abcd1234bls-----
`)
	got, err := keys.ParsePEMHeader(data)
	require.NoError(t, err)
	require.Equal(t, "abcd1234bls", got)
}

func TestParsePEMHeader_Wallet(t *testing.T) {
	data := []byte(`-----BEGIN PRIVATE KEY for klv1wallet9abcd-----
ZmFrZV9ib2R5...
-----END PRIVATE KEY for klv1wallet9abcd-----`)
	got, err := keys.ParsePEMHeader(data)
	require.NoError(t, err)
	require.Equal(t, "klv1wallet9abcd", got)
}

func TestParsePEMHeader_Malformed(t *testing.T) {
	_, err := keys.ParsePEMHeader([]byte("not a pem file"))
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestParsePEMHeader_Empty(t *testing.T) {
	_, err := keys.ParsePEMHeader(nil)
	require.Error(t, err)
}
