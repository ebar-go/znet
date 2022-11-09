package codec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeader_IsContentTypeJson(t *testing.T) {
	header := &Header{Options: OptionContentTypeJson}
	assert.Equal(t, OptionContentTypeJson, header.Options&OptionContentTypeJson)
}
