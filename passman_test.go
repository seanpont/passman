package main

import (
	_"fmt"
	_"io/ioutil"
	"bytes"
	"testing"
	"github.com/seanpont/assert"
)


func TestReadPasswordFile(t *testing.T) {
	assert := assert.Assert(t)
	// passing nil returns an empty list
	pwEntries, err := readPasswordFile(nil)
	assert.Nil(err)
	assert.Equal(pwEntries, make([]PasswordEntry, 0))
	
	// happy case: one entry
	pwEntries = make([]PasswordEntry, 0)
	pwEntries = append(pwEntries, PasswordEntry{"github", "asdfasdf"})
	buffer := new(bytes.Buffer)
	err = writePwFile(buffer, pwEntries)
	// fmt.Println(ioutil.ReadAll(&buffer))
	pwEntries, err = readPasswordFile(buffer)
	assert.Equal(len(pwEntries), 1)

	// two entries
	pwEntries = append(pwEntries, PasswordEntry{"google", "fdsfdsaf"})
	buffer = new(bytes.Buffer)
	err = writePwFile(buffer, pwEntries)
	assert.Nil(err)
	pwEntries, err = readPasswordFile(buffer)
	assert.Nil(err)
	assert.Equal(len(pwEntries), 2)

	// does not barf on unknown keys
	badJson := "[{\"Stuff\": \"Bland\"}]"
	pwEntries, err = readPasswordFile(bytes.NewBufferString(badJson))
	assert.Equal(len(pwEntries), 1)
	assert.Nil(err)

	// truly malformed returns empty array and error
	badJson = "asdfasdfg-\""
	pwEntries, err = readPasswordFile(bytes.NewBufferString(badJson))
	assert.Equal(len(pwEntries), 0)
	assert.NotNil(err)
}

