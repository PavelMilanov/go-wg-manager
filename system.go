package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

/*
Проверка операционной системы на совместимость.
*/
func initSystem() {
	_, err := exec.Command("bash", "-c", "cat /etc/os-release").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/*
Динамическое назначение приватных ip-адресов клиентам.
*/
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
	if len(configs) > 0 && label == configs[len(configs)-1].ClientLocalAddress {
		label = fmt.Sprintf("10.0.0.%d/24", lastindex)
	}
	return label
}

/*
Автопоиск интерфейса и ip для конфигурации сервера.
*/
func setServerParams() (string, string) {
	out, err := exec.Command("bash", "-c", "ip r").Output()
	if err != nil {
		panic(err)
	}
	var serverIp, serverIntf string
	defaultRoute := strings.Split(string(out), " ")[:5] // первая строка "default via 192.168.11.1 dev vlan601 proto static metric 404 ..."
	ip := defaultRoute[2]
	gate4 := net.ParseIP(ip)
	serverIntf = defaultRoute[4]
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, interf := range interfaces {
		// Список адресов для каждого сетевого интерфейса
		addrs, err := interf.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			data := addr.String()
			ip, ipnet, _ := net.ParseCIDR(data)
			if ipnet.Contains(gate4) {
				serverIp = ip.String()
			}
		}
	}
	return serverIp, serverIntf
}

/*
Генерация приватного и публичного ключей.
*/
func generateKeys() (string, string) {
	dir := os.TempDir()
	os.Chdir(dir)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
	privatekeyToFile, _ := os.ReadFile("privatekey")
	publickeyToFile, _ := os.ReadFile("publickey")
	privatekey := strings.TrimRight(string(privatekeyToFile), "\n")
	publickey := strings.TrimRight(string(publickeyToFile), "\n")
	defer os.RemoveAll(dir)
	return privatekey, publickey
}

/*
Просмотр статистики wg.
*/
func showPeers() {
	out, err := exec.Command("bash", "-c", "sudo wg show").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

/*
Перезапуск службы wg.
*/
func restartServer() {
	server := readServerConfigFile()
	command := fmt.Sprintf("sudo systemctl restart wg-quick@%s.service", server.Alias)
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
