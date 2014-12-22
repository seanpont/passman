package main

import (
	"bytes"
	"code.google.com/p/go.crypto/blowfish"
	"code.google.com/p/gopass"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/seanpont/gobro"
	"github.com/seanpont/gobro/commander"
	"github.com/seanpont/gobro/strarr"
	"io/ioutil"
	"os"
	"os/user"
	"sort"
	"strings"
)

// ===== CONFIGURATION =======================================================

type Config struct {
	PasswdFile string
}

func configFileName() string {
	u, err := user.Current()
	gobro.CheckErr(err)
	return u.HomeDir + "/.passman-config"
}

func loadConfig() (config Config, err error) {
	configFile, err := os.Open(configFileName())
	if err != nil {
		return
	}
	defer configFile.Close()
	err = json.NewDecoder(configFile).Decode(&config)
	return
}

func saveConfig(config Config) error {
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
	if config.PasswdFile == "" {
		fmt.Fprintln(os.Stderr, "The config file is missing or malformed.")
		os.Exit(1)
	}
	return config.PasswdFile
}

func loadServices(passwd string) (*Services, error) {
	base64Enc, err := ioutil.ReadFile(passwdFileName())
	if err != nil {
		return nil, err
	}

	enc := make([]byte, base64.StdEncoding.DecodedLen(len(base64Enc)))
	base64.StdEncoding.Decode(enc, base64Enc)

	blockCipher, err := blowfish.NewCipher([]byte(passwd))
	gobro.CheckErr(err)
	decrypted := make([]byte, len(enc)-8)
	cipher.NewCFBDecrypter(blockCipher, enc[:8]).XORKeyStream(decrypted, enc[8:])

	services := new(Services)
	err = json.NewDecoder(bytes.NewBuffer(decrypted)).Decode(services)
	if err != nil {
		return nil, errors.New("Invalid password")
	}
	return services, nil
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

	enc = append(iv, enc...)
	base64EncodedLen := base64.StdEncoding.EncodedLen(len(enc))
	base64Enc := make([]byte, base64EncodedLen)
	base64.StdEncoding.Encode(base64Enc, enc)

	err = ioutil.WriteFile(passwdFileName(), base64Enc, os.ModePerm)
	gobro.CheckErr(err)
}

// ===== HELPERS =============================================================

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

func initialize(args []string) {
	_, err := loadConfig()
	if err != nil {
		u, err := user.Current()
		gobro.CheckErr(err)
		saveConfig(Config{PasswdFile: u.HomeDir + "/.passman"})
	}
	services, err := loadServices(getPasswd())
	if err == nil && services != nil && services.Len() > 0 {
		overwrite, err := commander.Prompt("   Overwrite existing password file? (y/N): ")
		if err != nil || strings.ToLower(overwrite) != "y" {
			return
		}
	}
	saveServices(getPasswd(), new(Services))
	_, err = loadServices(getPasswd()) // just to test that we have saved it properly
	gobro.CheckErr(err, "Unable to read password file")
}

func configure(args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Config file not found. Please initialize first.")
		return
	}

	d := strarr.IndexOf(args, "-d")
	if d >= 0 && len(args) > d {
		config.PasswdFile = args[d+1]
	}
	saveConfig(config)

	fmt.Println(config.PasswdFile)
}

func add(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: passman add <service> [<options>]\n"+
			"  -g: Generate password\n  "+
			"    Additional options:\n"+
			"      l: include lowercase characters\n"+
			"      u: include uppercase characters\n"+
			"      n: include numbers\n"+
			"      c: include special characters\n"+
			"      w: generate a password using random words from the dictionary\n"+
			"      \\d+$: password must be at least this long."+
			"      example: add -glun24 will produce a password using, lowercase, characters,\n"+
			"      uppercase characters, and numbers, and will be 24 characters long"+
			"  -p: Enter password\n  "+
			"  -m: attach metadata\n")
		return
	}

	name := args[0]
	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	service := services.Get(name)
	service.Name = name

	generateParams, _ := strarr.FindMatchWithRegex(args, "-g.*")
	if generateParams != "" {
		service.Password = NewPasswordGenerator(generateParams).generate()
	}

	if strarr.Contains(args, "-p") {
		prompt := fmt.Sprintf("   Password for %s: ", service.Name)
		password, err := gopass.GetPass(prompt)
		gobro.CheckErr(err)
		service.Password = password
	}

	if strarr.Contains(args, "-m") {
		service.Meta, _ = commander.Prompt(fmt.Sprintf("   Meta for %s: ", service.Name))
	}

	services.Add(service)
	saveServices(getPasswd(), services)
}

func ls(args []string) {
	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	if len(args) > 0 {
		services.Services = services.Search(args[0])
	}
	for _, service := range services.Services {
		fmt.Printf("   %s\n", service.Name)
	}
}

func show(args []string) {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}

	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	for _, service := range services.Search(prefix) {
		fmt.Printf("   %20s:   %s   %s\n", service.Name, service.Password, service.Meta)
	}
	fmt.Println("")
}

func rm(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: passman rm <services>")
		return
	}
	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
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
	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	service := services.Get(args[0])
	if service.Name != "" {
		clipboard.WriteAll(service.Password)
	} else {
		fmt.Printf("'%s' not found\n", args[0])
	}
}

func changePassword(args []string) {
	if passwd == nil {
		pw, err := gopass.GetPass("   Current password: ")
		gobro.CheckErr(err)
		if pw == "" {
			fmt.Fprintln(os.Stderr, "Invalid password")
			return
		}
	}
	services, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	pw2, err := gopass.GetPass("   New Password: ")
	gobro.CheckErr(err)
	if pw2 == "" {
		fmt.Fprintln(os.Stderr, "Invalid password")
		return
	}
	pw3, err := gopass.GetPass("   Repeat Password: ")
	gobro.CheckErr(err)
	if pw3 == "" {
		fmt.Fprintln(os.Stderr, "Invalid password")
		return
	}
	if pw2 != pw3 {
		fmt.Fprintln(os.Stderr, "Passwords don't match")
		return
	}
	saveServices(pw2, services)
	passwd = &pw2
}

func exit(args []string) {
	fmt.Println("Goodbye")
	os.Exit(0)
}

func commandMap() *commander.CommandMap {
	cm := commander.NewCommandMap()
	cm.Add("config", configure, "Configure the password manager")
	cm.Add("init", initialize, "Initialize the password manager")
	cm.Add("add", add, "Add a service")
	cm.Add("ls", ls, "List services")
	cm.Add("show", show, "Show password for a service")
	cm.Add("rm", rm, "Remove a service")
	cm.Add("cp", cp, "Copy the password for a service to the clipboard")
	cm.Add("repass", changePassword, "Change the encryption password")
	cm.Add("exit", exit, "Exit")
	return cm
}

func interactive(args []string) {
	cm := commandMap()
	_, err := loadServices(getPasswd())
	gobro.CheckErr(err, "Password invalid")
	for {
		cmd, err := commander.Prompt("passman$ ")
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
	cm := commandMap()
	cm.Add("i", interactive, "Interactive mode (good for adding lots of passwords)")
	cm.Run(os.Args)
}
