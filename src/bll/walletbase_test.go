package bll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalletOutputSetLevel(t *testing.T) {
	for _, c := range []struct {
		input uint64
		level uint8
	}{
		{0, 0},
		{9, 0},
		{10, 1},
		{99, 1},
		{100, 2},
		{9999999999, 9},
		{10000000000, 10},
	} {
		w := &WalletOutput{Credits: c.input}
		w.SetLevel()
		assert.Equal(t, c.level, w.Level)
	}
}
