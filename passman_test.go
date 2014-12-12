package main

import (
	"github.com/seanpont/assert"
	"strings"
	"testing"
)

func TestServices(t *testing.T) {
	assert := assert.Assert(t)
	services := new(Services)
	assert.Equal(len(services.Services), 0)

	// Addition
	services.Add(Service{Name: "google.com", Password: "asdf1234", Meta: "Personal"})
	assert.Equal(len(services.Services), 1)
	assert.Equal(services.Services[0].Name, "google.com")

	// replacement
	services.Add(Service{Name: "google.com", Password: "lkjpoiu"})
	assert.Equal(len(services.Services), 1)
	assert.Equal(services.Services[0].Password, "lkjpoiu")

	// Addition
	services.Add(Service{Name: "facebook.com", Password: "1234"})
	assert.Equal(len(services.Services), 2)

	// another replacement
	services.Add(Service{Name: "google.com", Password: "asghasd"})
	assert.Equal(services.Get("google.com").Password, "asghasd")
	assert.Equal(len(services.Services), 2)

	services.Add(Service{Name: "golang.org"})
	goServices := services.Search("go")
	assert.Equal(len(goServices), 2)
	assert.True(goServices[0].Name[:2] == "go", "Service 0 begins with go")
	assert.True(goServices[1].Name[:2] == "go", "Service 1 begins with go")
	assert.Equal(len(services.Search("*")), 3)
}

func TestPasswordGeneration(t *testing.T) {
	assert := assert.Assert(t)
	generator := NewPasswordGenerator("lunc20")
	assert.Equal(generator.length, 20)
	password := generator.generate()
	assert.Equal(len(password), 20)
	assert.True(strings.ContainsAny(password, LOWERCASE), "missing characters")
	assert.True(strings.ContainsAny(password, UPPERCASE), "missing characters")
	assert.True(strings.ContainsAny(password, NUMBERS), "missing characters")
	assert.True(strings.ContainsAny(password, CHARACTERS), "missing characters")
}

func TestPasswordGenerationWithWords(t *testing.T) {
	assert := assert.Assert(t)
	generator := NewPasswordGenerator("w24")
	assert.Equal(generator.length, 24)
	assert.Equal(generator.words, true)
	password := generator.generate()
	assert.True(len(password) >= 24, "length")
}
