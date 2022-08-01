# building for Rapsberry Pi

```sh
 env GOOS=linux GOARCH=arm GOARM=5 go build -o forwarder
 ```

 # running
 * copy to homegear machine as /usr/local/bin/forwarder
 * turn on logging so the relevant messages show up
 * create service
 * enable service by running `systemctl enable mqtt-forwarder`

## service definiton

place this in `/etc/systemd/system/mqtt-forwarder.service`

```
[Unit]
Description=fixes homegear issues with homematic energy sensor
After=homegear.target
Wants=homegear.target
StartLimitIntervalSec=0

[Service]
User=root
WorkingDirectory=/tmp
ExecStart=/usr/local/bin/forwarder
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
