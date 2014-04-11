package main

import (
	"fmt"
	_"io/ioutil"
	"bytes"
	"testing"
	"github.com/seanpont/assert"
)

func TestPad(t *testing.T) {
	assert := assert.Assert(t)

	// len < n
	key := "abc"
	padded := pad(key, 4)
	assert.Equal(len(padded), 4)

	// len == n
	key = "asdf"
	padded = pad(key, 4)
	assert.Equal(padded, []byte(key))

	// len > n
	key = "asdfa"
	padded = pad(key, 4)
	assert.Equal(len(padded), 8)
}

func TestEncryptAndDecrypt(t *testing.T) {
	assert := assert.Assert(t)
	// encryption works
	key, plaintext := "secret key", "Some very important text"
	ciphertext := encrypt(plaintext, key)
	
	fmt.Printf("ciphertext: %s", ciphertext)
	plaintext2 := decrypt(ciphertext, key)

	assert.Equal(plaintext, plaintext2)
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

