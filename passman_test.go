package main

import (
	_ "bytes"
	_ "fmt"
	"github.com/seanpont/assert"
	_ "io/ioutil"
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
}
