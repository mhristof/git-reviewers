package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSustract(t *testing.T) {
	cases := []struct {
		name   string
		items  []string
		remove []string
		exp    []string
	}{
		{
			name:   "simple case",
			items:  []string{"1", "2", "3"},
			remove: []string{"1"},
			exp:    []string{"2", "3"},
		},
	}

	for _, test := range cases {
		assert.Equal(t, test.exp, Subtract(test.items, test.remove), test.name)
	}
}
