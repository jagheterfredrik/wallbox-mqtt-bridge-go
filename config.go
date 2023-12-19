package main

import (
    "time"

    "gopkg.in/ini.v1"
)

type WallboxConfig struct {
    MQTT struct {
        Host     string `ini:"host"`
        Port     int    `ini:"port"`
        Username string `ini:"username"`
        Password string `ini:"password"`
    } `ini:"mqtt"`

    Settings struct {
        PollingIntervalSeconds time.Duration `ini:"polling_interval_seconds"`
        DeviceName             string        `ini:"device_name"`
    } `ini:"settings"`
}

func LoadConfig(path string) *WallboxConfig {
    cfg, _ := ini.Load(path)

    var config WallboxConfig
    if err := cfg.MapTo(&config); err != nil {
        return nil
    }

    return &config
}
