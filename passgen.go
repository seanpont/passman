package main

import (
	"crypto/rand"
	"github.com/seanpont/gobro"
	"io/ioutil"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ===== PASSWORD GENERATOR ==================================================

const (
	LOWERCASE  = "abcdefghijklmnopqrstuvwxyz"
	UPPERCASE  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NUMBERS    = "1234567890"
	CHARACTERS = "!@#$%^&*?-_=+"
)

type PasswordGenerator struct {
	lowercase  bool
	uppercase  bool
	numbers    bool
	characters bool
	words      bool
	length     int
	dictionary []string
}

func NewPasswordGenerator(config string) *PasswordGenerator {
	regexp, err := regexp.Compile("\\d+$")
	gobro.CheckErr(err)
	passwordLength := 24 // default
	lengthBytes := regexp.Find([]byte(config))
	if len(lengthBytes) > 0 {
		length, err := strconv.Atoi(string(lengthBytes))
		if err == nil {
			passwordLength = length
		}
	}

	lowercase := strings.Contains(config, "l")
	uppercase := strings.Contains(config, "u")
	numbers := strings.Contains(config, "n")
	characters := strings.Contains(config, "c")
	words := strings.Contains(config, "w")

	if !(words || lowercase || uppercase || numbers || characters) {
		// Default values when no options specified
		lowercase = true
		uppercase = true
		numbers = true
	}

	generator := &PasswordGenerator{
		lowercase:  lowercase,
		uppercase:  uppercase,
		numbers:    numbers,
		characters: characters,
		words:      words,
		length:     passwordLength,
	}

	if words {
		generator.loadDictionary()
	} else {
		generator.loadCharacterTypes()
	}

	return generator
}

func (gen *PasswordGenerator) loadCharacterTypes() {
	dictionary := make([]string, 0)
	if gen.lowercase {
		dictionary = append(dictionary, strings.Split(LOWERCASE, "")...)
	}
	if gen.uppercase {
		dictionary = append(dictionary, strings.Split(UPPERCASE, "")...)
	}
	if gen.numbers {
		dictionary = append(dictionary, strings.Split(NUMBERS, "")...)
	}
	if gen.characters {
		dictionary = append(dictionary, strings.Split(CHARACTERS, "")...)
	}
	gen.dictionary = dictionary
}

func (gen *PasswordGenerator) loadDictionary() {
	dictFile, err := os.Open("/usr/share/dict/words")
	gobro.CheckErr(err)
	dictBytes, err := ioutil.ReadAll(dictFile)
	gobro.CheckErr(err)
	gen.dictionary = strings.Split(string(dictBytes), "\n")
}

func (gen *PasswordGenerator) generate() string {
	if gen.words {
		return gen.generateWithWords()
	} else {
		return gen.generateWithCharacters()
	}
}

func (gen *PasswordGenerator) generateWithWords() string {
	// use /usr/share/dict
	dictFile, err := os.Open("/usr/share/dict/words")
	gobro.CheckErr(err)
	dictBytes, err := ioutil.ReadAll(dictFile)
	gobro.CheckErr(err)
	words := strings.Split(string(dictBytes), "\n")

	password := ""
	maxSubscript := big.NewInt(int64(len(words)))
	for len(password) < gen.length {
		subscript, _ := rand.Int(rand.Reader, maxSubscript)
		password += strings.Title(words[int(subscript.Int64())])
	}

	return password
}

func (gen *PasswordGenerator) generateWithCharacters() string {
	password := ""
	maxSubscript := big.NewInt(int64(len(gen.dictionary)))
	for i := 0; i < gen.length; i++ {
		subscript, _ := rand.Int(rand.Reader, maxSubscript)
		password += gen.dictionary[int(subscript.Int64())]
	}

	// validate that it contains all required character types
	if gen.lowercase && !strings.ContainsAny(password, LOWERCASE) ||
		gen.uppercase && !strings.ContainsAny(password, UPPERCASE) ||
		gen.numbers && !strings.ContainsAny(password, NUMBERS) ||
		gen.characters && !strings.ContainsAny(password, CHARACTERS) {
		// try again!
		return gen.generateWithCharacters()
	}
	return password
}
