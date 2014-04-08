package main

import (
	"fmt"
	_"io/ioutil"
	"bytes"
	"testing"
	"github.com/seanpont/assert"
)

func TestPadKey(t *testing.T) {
	assert := assert.Assert(t)
	key := "this is a short key"
	assert.True(len(key) < 32, "bad length")
	padded := padKey(key)
	assert.Equal(len(padded), 32)

	// Longer keys are trimmed
	key = "this is a long key the length of which is over 32 bytes"
	assert.True(len(key) > 32, "long key not long enough")
	padded = padKey(key)
	assert.Equal(len(padded), 32)
}

func TestEncryptPassword(t *testing.T) {
	assert := assert.Assert(t)
	// encryption works
	salt, pw, key := "salt", "secret", "key"
	
	encrypted, _ := encryptPassword(salt, pw, key)
	fmt.Println(encrypted)
	assert.NotNil(encrypted)

}

func TestReadPasswordFile(t *testing.T) {
	assert := assert.Assert(t)
	// passing nil returns an empty list
	pwEntries, err := readPwFile(nil)
	assert.Nil(err)
	assert.Equal(pwEntries, make([]PasswordEntry, 0))
	
	// happy case: one entry
	pwEntries = make([]PasswordEntry, 0)
	pwEntries = append(pwEntries, PasswordEntry{"github", "asdfasdf"})
	buffer := new(bytes.Buffer)
	err = writePwFile(buffer, pwEntries)
	// fmt.Println(ioutil.ReadAll(&buffer))
	pwEntries, err = readPwFile(buffer)
	assert.Equal(len(pwEntries), 1)

	// two entries
	pwEntries = append(pwEntries, PasswordEntry{"google", "fdsfdsaf"})
	buffer = new(bytes.Buffer)
	err = writePwFile(buffer, pwEntries)
	assert.Nil(err)
	pwEntries, err = readPwFile(buffer)
	assert.Nil(err)
	assert.Equal(len(pwEntries), 2)

	// does not barf on unknown keys
	badJson := "[{\"Stuff\": \"Bland\"}]"
	pwEntries, err = readPwFile(bytes.NewBufferString(badJson))
	assert.Equal(len(pwEntries), 1)
	assert.Nil(err)

	// truly malformed returns empty array and error
	badJson = "asdfasdfg-\""
	pwEntries, err = readPwFile(bytes.NewBufferString(badJson))
	assert.Equal(len(pwEntries), 0)
	assert.NotNil(err)
}

