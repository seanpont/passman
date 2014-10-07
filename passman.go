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
	"github.com/atotto/clipboard"
	"github.com/seanpont/gobro"
	"io/ioutil"
	"os"
	"os/user"
	"sort"
	"strings"
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

// ===== SERVICES ============================================================

type Service struct {
	Name, Password, Meta string
}

type Services struct {
	Services []Service
}

// ----- sort.Interface ------------------------------------------------------

func (s *Services) Len() int {
	return len(s.Services)
}

func (s *Services) Less(i, j int) bool {
	return s.Services[i].Name < s.Services[j].Name
}

func (s *Services) Swap(i, j int) {
	s.Services[i], s.Services[j] = s.Services[j], s.Services[i]
}

// ----- OPERATIONS ----------------------------------------------------------

func (s *Services) IndexOf(name string) int {
	for i, service := range s.Services {
		if service.Name == name {
			return i
		}
	}
	return -1
}

func (s *Services) Get(name string) (service Service) {
	i := s.IndexOf(name)
	if i >= 0 {
		service = s.Services[i]
	}
	return
}

func (s *Services) Remove(name string) (service Service) {
	i := s.IndexOf(name)
	if i >= 0 {
		service = s.Services[i]
		s.Services = append(s.Services[:i], s.Services[i+1:]...)
	}
	return
}

func (s *Services) Add(service Service) {
	s.Remove(service.Name)
	s.Services = append(s.Services, service)
}

func (s *Services) Search(prefix string) []Service {
	prefix = strings.TrimSuffix(prefix, "*")
	matched := make([]Service, 0)
	for _, service := range s.Services {
		if strings.HasPrefix(service.Name, prefix) {
			matched = append(matched, service)
		}
	}
	return matched
}

// ===== PASSWORD FILE =======================================================

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
	sort.Sort(services)
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

// ===== HELPERS =============================================================

func generatePassword() string {
	randBytes := make([]byte, 64)
	rand.Read(randBytes)
	base64Buffer := new(bytes.Buffer)
	base64.NewEncoder(base64.StdEncoding, base64Buffer).Write(randBytes)
	return string(base64Buffer.Bytes()[:24])
}

var passwd *string

func getPasswd() string {
	if passwd == nil {
		pw, err := gopass.GetPass("Password: ")
		gobro.CheckErr(err)
		if pw == "" {
			fmt.Fprintln(os.Stderr, "Invalid password")
			os.Exit(1)
		}
		passwd = &pw
	}
	return *passwd
}

// ===== COMMANDS ============================================================

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

func initialize(args []string) {
	saveServices(getPasswd(), new(Services))
}

func add(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: passman add <service> [<options>]\n  -g: Generate password")
		return
	}

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

	services := loadServices(getPasswd())
	services.Add(*service)
	saveServices(getPasswd(), services)
}

func ls(args []string) {
	services := loadServices(getPasswd())
	if len(args) > 0 {
		services.Services = services.Search(args[0])
	}
	for _, service := range services.Services {
		fmt.Printf("   %s\n", service.Name)
	}
}

func show(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: passman show <service>")
		return
	}

	services := loadServices(getPasswd())
	for _, service := range services.Search(args[0]) {
		fmt.Printf("   %-20s:   %-60s %s\n", service.Name, service.Password, service.Meta)
	}
	fmt.Println("")
}

func rm(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: passman rm <services>")
		return
	}
	services := loadServices(getPasswd())
	for _, name := range args {
		services.Remove(name)
	}
	saveServices(getPasswd(), services)
}

func cp(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: passman cp <service>")
		return
	}
	services := loadServices(getPasswd())
	service := services.Get(args[0])
	if service.Name != "" {
		clipboard.WriteAll(service.Password)
	} else {
		fmt.Printf("'%s' not found\n", args[0])
	}
}

func interactive(args []string) {
	cm := gobro.NewCommandMap()
	cm.Add("config", configure, "Configure the password manager")
	cm.Add("init", initialize, "Initialize the password manager")
	cm.Add("add", add, "Add a service")
	cm.Add("ls", ls, "List services")
	cm.Add("show", show, "Show password for a service")
	cm.Add("rm", rm, "Remove a service")
	cm.Add("cp", cp, "Copy the password for a service to the clipboard")

	loadServices(getPasswd())
	for {
		cmd, err := gobro.Prompt("passman$ ")
		if err != nil {
			return
		}
		if cmd == "" {
			continue
		}
		args := make([]string, 0, 3)
		args = append(args, "passman")
		args = append(args, strings.Split(cmd, " ")...)
		cm.Run(args)
	}
}

func main() {
	cm := gobro.NewCommandMap()
	cm.Add("config", configure, "Configure the password manager")
	cm.Add("init", initialize, "Initialize the password manager")
	cm.Add("add", add, "Add a service")
	cm.Add("ls", ls, "List services")
	cm.Add("show", show, "Show password for a service")
	cm.Add("rm", rm, "Remove a service")
	cm.Add("cp", cp, "Copy the password for a service to the clipboard")
	cm.Add("i", interactive, "Interactive mode (good for adding lots of passwords)")
	cm.Run(os.Args)
}
