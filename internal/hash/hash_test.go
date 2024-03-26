package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	actual := Encode([]byte("message"), []byte("key"))
	expected := `bp7ym3X//Ft6uuUn1Y/a2y/kLnIZARl2kXNDBl9Y7Uo=`
	assert.Equal(t, actual, expected)
}
