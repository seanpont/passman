package main

import (
	"bytes"
	"code.google.com/p/go.crypto/blowfish"
	"code.google.com/p/gopass"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/seanpont/gobro"
	"io/ioutil"
	"os"
	"os/user"
	_ "strings"
)

// ===== CONFIGURATION =======================================================

type Config struct {
	PasswdDir string
}

func configFileName() string {
	u, err := user.Current()
	gobro.CheckErr(err)
	return u.HomeDir + "/.passman-config"
}

func loadConfig() (*Config, error) {
	configFile, err := os.Open(configFileName())
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	return &config, err
}

func saveConfig(config *Config) error {
	configFile, err := os.Create(configFileName())
	if err != nil {
		return err
	}
	defer configFile.Close()
	json.NewEncoder(configFile).Encode(config)
	return nil
}

func configure(args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Config file not found")
		config = new(Config)
	}

	d := gobro.IndexOf(args, "-d")
	if d >= 0 && len(args) > d {
		config.PasswdDir = args[d+1]
	}
	saveConfig(config)
}

// ===== PASSWORD MANAGEMENT =================================================

type Service struct {
	Name, Password, Meta string
}

type Services struct {
	Services []Service
}

func (s *Services) IndexOf(name string) int {
	for i, service := range s.Services {
		if service.Name == name {
			return i
		}
	}
	return -1
}

func (s *Services) Get(name string) (service *Service) {
	i := s.IndexOf(name)
	if i >= 0 {
		service = &s.Services[i]
	}
	return
}

func (s *Services) Remove(name string) (service *Service) {
	i := s.IndexOf(name)
	if i >= 0 {
		service = &s.Services[i]
		s.Services = append(s.Services[:i], s.Services[i+1:]...)
	}
	return
}

func (s *Services) Put(service *Service) {
	s.Remove(service.Name)
	s.Services = append(s.Services, *service)
}

func passwdFileName() string {
	config, err := loadConfig()
	gobro.CheckErr(err, "Unable to load config file")
	return config.PasswdDir + "/.passman"
}

func loadServices(passwd string) *Services {
	encrypted, err := ioutil.ReadFile(passwdFileName())
	gobro.CheckErr(err)

	blockCipher, err := blowfish.NewCipher([]byte(passwd))
	gobro.CheckErr(err)
	decrypted := make([]byte, len(encrypted)-8)
	cipher.NewCFBDecrypter(blockCipher, encrypted[:8]).XORKeyStream(decrypted, encrypted[8:])

	services := new(Services)
	err = json.NewDecoder(bytes.NewBuffer(decrypted)).Decode(services)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid password")
		os.Exit(1)
	}
	return services
}

func saveServices(passwd string, services *Services) {
	decryptedBuffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(decryptedBuffer).Encode(services)
	gobro.CheckErr(err)

	blockCipher, err := blowfish.NewCipher([]byte(passwd))
	gobro.CheckErr(err)
	iv := make([]byte, 8)
	_, err = rand.Read(iv)
	gobro.CheckErr(err)

	enc := decryptedBuffer.Bytes()
	cipher.NewCFBEncrypter(blockCipher, iv).XORKeyStream(enc, enc)
	buff := bytes.NewBuffer(iv)
	_, err = buff.Write(enc)
	gobro.CheckErr(err)

	err = ioutil.WriteFile(passwdFileName(), buff.Bytes(), os.ModePerm)
	gobro.CheckErr(err)
}

func generatePassword() string {
	randBytes := make([]byte, 64)
	rand.Read(randBytes)
	base64Buffer := new(bytes.Buffer)
	base64.NewEncoder(base64.StdEncoding, base64Buffer).Write(randBytes)
	return string(base64Buffer.Bytes()[:24])
}

func add(args []string) {
	rootPasswd := getPasswd()

	service := new(Service)
	service.Name = args[0]

	if gobro.Contains(args, "-g") {
		service.Password = generatePassword()
	} else {
		prompt := fmt.Sprintf("Password for %s: ", service.Name)
		password, err := gopass.GetPass(prompt)
		gobro.CheckErr(err)
		service.Password = password
	}

	services := loadServices(rootPasswd)
	services.Put(service)
	saveServices(rootPasswd, services)
}

func initialize(args []string) {
	saveServices(getPasswd(), new(Services))
}

func ls(args []string) {
	services := loadServices(getPasswd())
	for _, service := range services.Services {
		fmt.Printf("   %15s:   %-32s   %s\n", service.Name, service.Password, service.Meta)
	}
}

func getPasswd() string {
	passwd, err := gopass.GetPass("Password: ")
	gobro.CheckErr(err)
	if passwd == "" {
		fmt.Fprintln(os.Stderr, "Invalid password")
		os.Exit(1)
	}
	return passwd
}

func rm(args []string) {
	gobro.CheckArgs(args, 1, "Usage: passman rm <service>")
	passwd := getPasswd()
	services := loadServices(passwd)
	services.Remove(args[0])
	saveServices(passwd, services)
}

func main() {
	gobro.NewCommandMap(configure, add, initialize, ls, rm).Run(os.Args)
}
