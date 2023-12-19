package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var service = `[Unit]
Description=MQTT Bridge
After=network.target
Requires=mysqld.service
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=/home/root/mqtt-bridge/bridge /home/root/mqtt-bridge/bridge.ini

[Install]
WantedBy=multi-user.target
`

func AskConfirmOrNew(field *string, name string) {
	fmt.Printf("%s (%s): ", name, *field)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if len(input) > 0 {
		*field = input
	}
}

func AskConfirmOrNewInt(field *int, name string) {
	fmt.Printf("%s (%d): ", name, *field)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if len(input) > 0 {
		*field, _ = strconv.Atoi(input)
	}
}

func InstallService() {
	ioutil.WriteFile("/lib/systemd/system/mqtt-bridge.service", []byte(service), 0644)
	var cmd *exec.Cmd
	cmd = exec.Command("systemctl", "daemon-reload")
	cmd.Run()
	cmd = exec.Command("systemctl", "enable", "mqtt-bridge")
	cmd.Run()
	cmd = exec.Command("systemctl", "restart", "mqtt-bridge")
	cmd.Run()
}

func RunConfigTui() {
	config := WallboxConfig{}
	config.MQTT.Host = "127.0.0.1"
	config.MQTT.Port = 1883
	config.MQTT.Username = ""
	config.MQTT.Password = ""
	config.Settings.PollingIntervalSeconds = 1
	config.Settings.DeviceName = "Wallbox"

	AskConfirmOrNew(&config.MQTT.Host, "MQTT Host")
	AskConfirmOrNewInt(&config.MQTT.Port, "MQTT Port")
	AskConfirmOrNew(&config.MQTT.Username, "MQTT Username")
	AskConfirmOrNew(&config.MQTT.Password, "MQTT Password")
	AskConfirmOrNewInt(&config.Settings.PollingIntervalSeconds, "Polling interval")
	AskConfirmOrNew(&config.Settings.DeviceName, "Device name")

	config.SaveTo("bridge.ini")

	InstallService()
}
