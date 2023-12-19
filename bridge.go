package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func getEntitiesConfig(w *Wallbox) map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"added_energy": {
			"component": "sensor",
			"getter":    func() interface{} { return w.Data.RedisState.ScheduleEnergy },
			"config": map[string]interface{}{
				"name":                        "Added energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": 1,
			},
		},
		"added_range": {
			"component": "sensor",
			"getter":    func() interface{} { return w.Data.SQL.AddedRange },
			"config": map[string]interface{}{
				"name":                        "Added range",
				"device_class":                "distance",
				"unit_of_measurement":         "km",
				"state_class":                 "total",
				"suggested_display_precision": 1,
				"icon":                        "mdi:map-marker-distance",
			},
		},
		"cable_connected": {
			"component": "binary_sensor",
			"getter":    w.GetCableConnected,
			"config": map[string]interface{}{
				"name":         "Cable connected",
				"payload_on":   1,
				"payload_off":  0,
				"icon":         "mdi:ev-plug-type1",
				"device_class": "plug",
			},
		},
		"charging_enable": {
			"component": "switch",
			"setter":    w.SetChargingEnable,
			"getter":    func() interface{} { return w.Data.SQL.ChargingEnable },
			"config": map[string]interface{}{
				"name":          "Charging enable",
				"payload_on":    1,
				"payload_off":   0,
				"command_topic": "~/set",
				"icon":          "mdi:ev-station",
			},
		},
		"charging_power": {
			"component": "sensor",
			"getter": func() interface{} {
				return w.Data.RedisM2W.Line1Power + w.Data.RedisM2W.Line2Power + w.Data.RedisM2W.Line3Power
			},
			"config": map[string]interface{}{
				"name":                        "Charging power",
				"device_class":                "power",
				"unit_of_measurement":         "W",
				"state_class":                 "total",
				"suggested_display_precision": 1,
			},
		},
		"cumulative_added_energy": {
			"component": "sensor",
			"getter":    func() interface{} { return w.Data.SQL.CumulativeAddedEnergy },
			"config": map[string]interface{}{
				"name":                        "Cumulative added energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total_increasing",
				"suggested_display_precision": 1,
			},
		},
		"halo_brightness": {
			"component": "number",
			"setter":    w.SetHaloBrightness,
			"getter":    func() interface{} { return w.Data.SQL.HaloBrightness },
			"config": map[string]interface{}{
				"name":                "Halo Brightness",
				"command_topic":       "~/set",
				"min":                 0,
				"max":                 100,
				"icon":                "mdi:brightness-percent",
				"unit_of_measurement": "%",
				"entity_category":     "config",
			},
		},
		"lock": {
			"component": "lock",
			"setter":    w.SetLocked,
			"getter":    func() interface{} { return w.Data.SQL.Lock },
			"config": map[string]interface{}{
				"name":           "Lock",
				"payload_lock":   1,
				"payload_unlock": 0,
				"state_locked":   1,
				"state_unlocked": 0,
				"command_topic":  "~/set",
			},
		},
		"max_charging_current": {
			"component": "number",
			"setter":    w.SetMaxChargingCurrent,
			"getter":    func() interface{} { return w.Data.SQL.MaxChargingCurrent },
			"config": map[string]interface{}{
				"name":                "Max charging current",
				"command_topic":       "~/set",
				"min":                 6,
				"max":                 w.GetAvailableCurrent(),
				"unit_of_measurement": "A",
				"device_class":        "current",
			},
		},
		"status": {
			"component": "sensor",
			"getter":    w.GetEffectiveStatus,
			"config": map[string]interface{}{
				"name": "Status",
			},
		},
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	panic("Connection to MQTT lost")
}

func main() {
	c := LoadConfig(os.Args[1])
	w := NewWallbox()
	w.UpdateCache()

	serialNumber := w.GetSerialNumber()
	entityConfig := getEntitiesConfig(w)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", c.MQTT.Host, c.MQTT.Port))
	opts.SetUsername(c.MQTT.Username)
	opts.SetPassword(c.MQTT.Password)
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topicPrefix := "wallbox_" + serialNumber

	for key, val := range entityConfig {
		component := val["component"].(string)
		uid := serialNumber + "_" + key
		config := map[string]interface{}{
			"~":           topicPrefix + "/" + key,
			"state_topic": "~/state",
			"unique_id":   uid,
			"device": map[string]interface{}{
				"identifiers": serialNumber,
				"name":        c.Settings.DeviceName,
			},
		}
		for k, v := range val["config"].(map[string]interface{}) {
			config[k] = v
		}
		jsonPayload, _ := json.Marshal(config)
		token := client.Publish("homeassistant/"+component+"/"+uid+"/config", 1, true, jsonPayload)
		token.Wait()
	}

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		field := strings.Split(msg.Topic(), "/")[1]
		payload := string(msg.Payload())
		setter := entityConfig[field]["setter"].(func(string))
		fmt.Println("Setting", field, payload)
		setter(payload)
	}

	topic := topicPrefix + "/+/set"
	client.Subscribe(topic, 1, messageHandler)

	ticker := time.NewTicker(c.Settings.PollingIntervalSeconds * time.Second)
	defer ticker.Stop()

	published := make(map[string]interface{})
	rateLimiter := map[string]*DeltaRateLimit{
		"charging_power": NewDeltaRateLimit(10, 100),
		"added_energy":   NewDeltaRateLimit(10, 50),
	}

	for {
		select {
		case <-ticker.C:
			w.UpdateCache()
			for key, val := range entityConfig {
				payload := val["getter"].(func() interface{})()
				bytePayload := []byte(fmt.Sprint(payload))
				if published[key] != payload {
					if rate, ok := rateLimiter[key]; ok && !rate.Allow(payload.(float64)) {
						continue
					}
					fmt.Println("Publishing: ", key, payload)
					token := client.Publish(topicPrefix+"/"+key+"/state", 1, true, bytePayload)
					token.Wait()
					published[key] = payload
				}
			}
		case <-interrupt():
			fmt.Println("Interrupted. Exiting...")
			client.Disconnect(250)
			os.Exit(0)
		}
	}

}

func interrupt() <-chan os.Signal {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	return interrupt
}
