package main

import (
	_"fmt"
	"testing"
	"github.com/seanpont/assert"
)


func TestReadPasswordFile(t *testing.T) {
	assert := assert.Assert(t)
	// passing nil returns an empty list
	pwEntries, err := readPasswordFile(nil)
	assert.Nil(err, "error not null")
	assert.Equal(pwEntries, make([]PasswordEntry, 0))


}

