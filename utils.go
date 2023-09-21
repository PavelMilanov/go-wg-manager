package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

func generateKeys() (string, string) {
	/*
		Генерация приватного и публичного ключей.
	*/
	dir := os.TempDir()
	os.Chdir(dir)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
	privatekey, _ := os.ReadFile("privatekey")
	publickey, _ := os.ReadFile("publickey")
	defer os.RemoveAll(dir)
	return string(privatekey), string(publickey)
}

func readServerConfigFile() *WgServerConfig {
	files, _ := os.ReadDir(WG_MANAGER_DIR)
	config := &WgServerConfig{}
	for _, file := range files {
		content, err := os.ReadFile(WG_MANAGER_DIR + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
	}
	return config
}

func readClientConfigFiles() []*UserConfig {
	files, _ := os.ReadDir(USERS_CONFIG_DIR)
	config := &UserConfig{}
	var configs []*UserConfig
	for _, file := range files {
		content, err := os.ReadFile(USERS_CONFIG_DIR + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
		configs = append(configs, config)
	}
	return configs
}

func addUSer() {
	var alias string
	fmt.Println("Enter client description:")
	alias_value, _ := fmt.Scanf("%s", &alias)
	if alias_value == 0 {
		os.Exit(1)
	}
	clientPrivKey, clientPubKey := generateKeys()
	clientip := setClientIp()
	server := readServerConfigFile()
	config := UserConfig{
		ClientPrivateKey:   clientPrivKey,
		ClientPublicKey:    clientPubKey,
		ClientLocalAddress: clientip,
		ServerPublicKey:    server.ServerPublicKey,
		ServerIp:           server.PublicAddress,
		ServerPort:         server.ListenPort,
	}
	clientFile := fmt.Sprintf("%s/%s.conf", USERS_DIR, alias)
	templ, err := template.ParseFiles("./client_template.conf")
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	config.addConfigUser(alias)
	defer file.Close()
}

func setClientIp() string {
	configs := readClientConfigFiles()
	label := "10.0.0.2/24"
	var lastindex = 3 // так как первый ip 10.0.0.(2)
	for index, config := range configs {
		if label <= config.ClientLocalAddress {
			label = fmt.Sprintf("10.0.0.%d/24", index+2)
		}
		lastindex += index
	}
	// если нет пропущенных адресов, выдаем следующий по списку
	if len(configs) > 1 && label == configs[len(configs)-1].ClientLocalAddress {
		label = fmt.Sprintf("10.0.0.%d/24", lastindex)
	}
	return label
}

func installServer() {
	/*
		Основаня логика установки WG Server.
	*/
	updatePackage()
	installWgServer()
	os.Mkdir(WG_MANAGER_DIR, 0666)
	privKey, pubKey := generateKeys()
	configureServer(privKey, pubKey)
}
func updatePackage() {
	/*
		Обновление пакетов deb.
	*/
	fmt.Println("Updating packages...")
	cmd := exec.Command("apt", "update", "-y")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func installWgServer() {
	/*
		Установка пакета wireguard.
	*/
	fmt.Println("Installing WireGuard Server...")
	cmd := exec.Command("apt", "install", "-y", "wireguard")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func configureServer(priv string, pub string) {
	/*
		Создание шаблона конфигурационного файла сервера.
	*/
	var (
		private_addr string
		port         int
		intf         string
		alias        string
		public_addr  string
	)
	fmt.Println("Enter private network: 10.0.0.1/24")
	private_addr_value, _ := fmt.Scanf("%s\r", &private_addr)
	if private_addr_value == 0 {
		private_addr = "10.0.0.1/24"
	} else {
		isValid, _ := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}`, private_addr)
		if !isValid {
			fmt.Println("Enter valid value. Example: 10.0.0.1/24")
			os.Exit(1)
		}
	}
	fmt.Println("Enter listen port: 51830")
	port_value, _ := fmt.Scanf("%d\r", &port)
	if port_value == 0 {
		port = 51830
	}
	fmt.Println("Enter NAT-interface:")
	intf_value, _ := fmt.Scanf("%s\r", &intf)
	if intf_value == 0 {
		fmt.Println("Enter NAT-interface")
		os.Exit(1)
	}
	fmt.Println("Enter IP-address:")
	public_addr_value, _ := fmt.Scanf("%s\r", &public_addr)
	if public_addr_value == 0 {
		fmt.Println("Enter IP-address")
		os.Exit(1)
	} else {
		isValid, _ := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, public_addr)
		if !isValid {
			fmt.Println("Enter valid value. Example: 100.0.0.1")
			os.Exit(1)
		}
	}
	fmt.Println("Enter alias: 'wg0'")
	alias_value, _ := fmt.Scanf("%s\r", &alias)
	if alias_value == 0 {
		alias = "wg0"
	}
	config := WgServerConfig{
		ServerPrivateKey: priv,
		ServerPublicKey:  pub,
		LocalAddress:     private_addr,
		PublicAddress:    public_addr,
		ListenPort:       port,
		Eth:              intf,
		Alias:            alias,
	}
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, alias)
	templ, err := template.ParseFiles("./wg_template.conf")
	file, err := os.OpenFile(serverFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	config.createServerConfigFile()
	defer file.Close()
}
