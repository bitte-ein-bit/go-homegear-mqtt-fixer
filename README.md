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
Description=mqtt-forwarder - fixes homegear issues with homematic energy sensor
After=network.target

[Service]
User=root
WorkingDirectory=/tmp
ExecStart=/usr/local/bin/forwarder
Restart=always

[Install]
WantedBy=multi-user.target
```