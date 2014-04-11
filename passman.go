package main

import (
	"code.google.com/p/gopass"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/seanpont/ergo"
	"io"
	"io/ioutil"
	_ "os"
	"os/user"
)

const KEY_LENGTH = 32

func main() {

	currentUser, err := user.Current()
	ergo.CheckNil(err)
	filename := currentUser.HomeDir + "/.passman"
	fmt.Println("filename: " + filename)

	init := flag.Bool("init", false, "Create a new .passman file")
	addService := flag.String("add", "", "Add a password")
	addServicePassword := flag.String("-p", "", "password to add")
	getService := flag.String("get", "", "Get a password")

	flag.Parse()

	if !*init && *addService == "" && *getService == "" {
		flag.PrintDefaults()
		return
	}

	pw := promptForPassword()
	pwEntries := load(filename, pw)

	if *init {
		save(filename, pwEntries, pw)
		return
	}

	if *addService != "" {
		servicePassword := *addServicePassword
		if servicePassword == "" {
			servicePassword = promptForPassword()
		}
		pwEntries[*addService] = servicePassword
		fmt.Println(pwEntries)
		save(filename, pwEntries, pw)
		return
	}

	if *getService != "" {
		fmt.Println(pwEntries[*getService])
		return
	}
}

type PasswordEntry struct {
	Service, Password, Meta string
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

func load(filename, pw string) map[string]string {
	return decode(decrypt(readFromDisk(filename), pw))
}

func save(filename string, pwEntries map[string]string, pw string) {
	writeToDisk(filename, encrypt(encode(pwEntries), pw))
}

func encode(data map[string]string) []byte {
	if data == nil {
		data = make(map[string]string)
	}
	j, err := json.Marshal(data)
	ergo.CheckNil(err)
	return j
}

func decode(data []byte) map[string]string {
	m := make(map[string]string)
	json.Unmarshal(data, &m)
	return m
}

func writeToDisk(filename string, data []byte) {
	err := ioutil.WriteFile(filename, data, 0644)
	ergo.CheckNil(err)
}

func readFromDisk(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	return data
}

// Returns a []byte padded to a multiple of the length specified
func pad(key []byte, n int) []byte {
	paddingRequired := (n - (len(key) % n)) % n
	padding := make([]byte, paddingRequired)
	return append(key, padding...)
}

func encrypt(plaintext []byte, key string) []byte {
	paddedKey := pad([]byte(key), 32)[:32]

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		panic(err)
	}

	plaintext = pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

func decrypt(ciphertext []byte, key string) []byte {
	paddedKey := pad([]byte(key), KEY_LENGTH)[:KEY_LENGTH]

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		panic(err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}
