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
	s1 := Service{Name: "google.com", Password: "asdf1234", Meta: "Personal"}
	services.Put(&s1)
	assert.Equal(len(services.Services), 1)
	assert.Equal(services.Services[0].Name, "google.com")

	// replacement
	s2 := Service{Name: "google.com", Password: "lkjpoiu"}
	services.Put(&s2)
	assert.Equal(len(services.Services), 1)
	assert.Equal(services.Services[0].Password, s2.Password)

	// Addition
	s3 := Service{Name: "facebook.com", Password: "1234"}
	services.Put(&s3)

	// another replacement
	s1.Password = "84824"
	services.Put(&s1)
	assert.Equal(services.Get("google.com").Password, s1.Password)
	assert.Equal(len(services.Services), 2)
}
