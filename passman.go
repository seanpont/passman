package main

import (
	"code.google.com/p/gopass"
	"crypto/aes"
	_ "flag"
	"fmt"
	"io"
	"io/ioutil"
	"encoding/json"
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
		addPassword(service, servicePassword, promptForPassword())
	} else {
		getPassword(firstArg)
	}
}

func printHelp() {
	// write man.txt to console
	manual, err := ioutil.ReadFile("man.txt")
	ergo.CheckNil(err)	
	fmt.Println(string(manual))
}

func promptForPassword() (password string) {
	password, err := gopass.GetPass("Password: ")
	ergo.CheckNil(err)
	return
}

type PasswordEntry struct {
	Service, Password string
}

func readPwFile(reader io.Reader) ([]PasswordEntry, error) {
	pwEntries := make([]PasswordEntry, 0)
	if reader == nil { return pwEntries, nil }
	err := json.NewDecoder(reader).Decode(&pwEntries)	
	return pwEntries, err	
}

func writePwFile(writer io.Writer, pwEntries []PasswordEntry) error {
	return json.NewEncoder(writer).Encode(pwEntries)
}

func addPassword(service, servicePassword, password string) {
	fmt.Println("adding a password")
	encryptedPw, _ := encryptPassword(service, servicePassword, password)
	fmt.Printf("Encrypted password: %v\n", encryptedPw)
}

func getPassword(service string) {
	fmt.Printf("getting password for service: %s\n", service)
}

// Returns a []byte of length 32
func padKey(key string) []byte {
	padded := []byte(key)
	if len(padded) >= 32 {
		return padded[:32]
	}
	padding := make([]byte, 32-len(padded))
	return append(padded, padding...)
}

func encryptPassword(salt, password, encryptionKey string) ([]byte, error) {
	paddedKey := padKey(encryptionKey)
	
	_, err := aes.NewCipher(paddedKey)
	if err != nil { return nil, err }
	
	return nil, nil
}




