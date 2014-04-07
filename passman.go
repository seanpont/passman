package main

import (
	"code.google.com/p/gopass"
	_ "crypto/aes"
	_ "flag"
	"fmt"
	"io"
	"io/ioutil"
	_"log"
	"os"
	"github.com/seanpont/ergo"
)

const FILE_NAME="~/.passman"

func main() {

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	firstArg := os.Args[1]

	if firstArg == "add" {
		if len(os.Args) < 4 {
			printHelp()
			return
		}
		service, servicePassword := os.Args[2], os.Args[3]
		addPassword(service, servicePassword)
	} else {
		getPassword(firstArg)
	}
}

func printHelp() {
	// write man.txt to console
	manual, err := ioutil.ReadFile("man.txt")
	ergo.Check(err)	
	fmt.Println(string(manual))
}

func promptForPassword() (password string) {
	password, err := gopass.GetPass("Password: ")
	ergo.Check(err)
	return
}

type PasswordEntry struct {
	service, password string
}

func readPasswordFile(reader io.Reader) ([]PasswordEntry, error) {
	return nil, nil	
}

func addPassword(service, servicePassword string) {
	fmt.Println("adding some passwords")
}

func getPassword(service string) {
	fmt.Printf("getting password for service: %s\n", service)
}
