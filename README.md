# MQTT Bridge for Wallbox

Rewrite in go to support new Wallbox without Python.

## Install
1. Get [the appropriate release](https://github.com/jagheterfredrik/wallbox-mqtt-bridge-go/releases) (or make it yourself, using ./make.sh)
2. Upload the binary as `~/mqtt-bridge/bridge`
3. Upload `bridge.ini` and `mqtt-bridge.service` to `~/mqtt-bridge/
4. Symlink the service: `ln -s /home/root/mqtt-bridge/mqtt-bridge.service /lib/systemd/system/mqtt-bridge.service`
5. Enable and start: `systemctl enable mqtt-bridge`, `systemctl start mqtt-bridge`
