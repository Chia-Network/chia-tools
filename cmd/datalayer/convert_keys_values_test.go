package datalayer

import (
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertFormats(t *testing.T) {
	testCases := []struct {
		name       string
		input      types.Bytes
		expect     string
		fromFormat string
		toFormat   string
	}{
		{
			name:       "Hex to UTF8",
			input:      types.Bytes("0x7631"),
			expect:     "v1",
			fromFormat: "hex",
			toFormat:   "utf8",
		},
		{
			name:       "UTF8 to Hex",
			input:      types.Bytes("v1"),
			expect:     "0x7631",
			fromFormat: "utf8",
			toFormat:   "hex",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := convertFormat(tc.input, tc.fromFormat, tc.toFormat)
			require.NoError(t, err)
			require.Equal(t, tc.expect, actual)
		})
	}
}
