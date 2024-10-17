package protocol

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodc(t *testing.T) {
	headers := make(Headers, 2)
	headers["method"] = "Hello.hi"
	headers["timeout"] = "1000"

	bs := headers.Marshal()

	hs := make(Headers, 2)
	hs.Unmarshal(bs)
	fmt.Printf("headers : %s", hs)
	assert.Equal(t, headers["method"], "Hello.hi")
}

func TestCodcEmpty(t *testing.T) {
	headers := make(Headers, 2)
	bs := headers.Marshal()

	hs := make(Headers, 2)
	hs.Unmarshal(bs)

	assert.Equal(t, len(bs), 0)
	assert.Equal(t, len(hs), 0)
}
