#!/bin/bash
arch=$(uname -m)
if [ "$arch" == "armv7l" ]; then
    curl -L --create-dirs -o ~/mqtt-bridge/bridge https://github.com/jagheterfredrik/wallbox-mqtt-bridge-go/releases/download/configuration-tool/bridge-armhf
elif [ "$arch" == "aarch64" ]; then
    curl -L --create-dirs -o ~/mqtt-bridge/bridge https://github.com/jagheterfredrik/wallbox-mqtt-bridge-go/releases/download/configuration-tool/bridge-arm64
else
    echo "Unknown architecture $arch"
fi
cd ~/mqtt-bridge/
chmod +x bridge
./bridge --config
