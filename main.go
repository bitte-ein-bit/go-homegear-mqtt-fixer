package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hpcloud/tail"
)

const (
	mqtt_topic_prefix = "homegear/1234-5678-9abc/plain"
	layout            = "01/02/06 15:04:05"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func publish(client mqtt.Client, topic string, payload float32) {
	fmt.Printf("%s -> %0.6f\n", topic, payload)
	t := client.Publish(topic, 0, false, fmt.Sprintf("%0.6f", payload))
	go func() {
		_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
		if t.Error() != nil {
			fmt.Printf("Error sending: %s\n", t.Error()) // Use your preferred logging technique (or just fmt.Printf)
		}
	}()
}

func connect() mqtt.Client {
	var broker = "docker"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	// opts.SetUsername("emqx")
	// opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

func main() {
	t, err := tail.TailFile("/var/log/homegear/homegear.log", tail.Config{
		Follow: true,
		ReOpen: true})
	if err != nil {
		panic(fmt.Errorf("cannot open logfile: %s", err))
	}

	client := connect()

	r, _ := regexp.Compile("([0-9/ :]+)\\.[0-9]+ Module HomeMatic BidCoS: Info: (IEC_ENERGY_COUNTER|IEC_POWER) on channel ([1|2]) of HomeMatic BidCoS peer ([0-9]+) with serial number [A-Za-z0-9]+ was set to 0x([A-F0-9]+)")

	for line := range t.Lines {
		// fmt.Println(line.Text)
		m := r.FindStringSubmatch(line.Text)
		if len(m) > 0 {
			timestamp := m[1]
			peer := m[4]
			channel := m[3]
			name := m[2]
			val := m[5]
			if peer == "1" && channel == "2" && name == "IEC_POWER" {
				continue
			}
			t, _ := time.Parse(layout, timestamp)
			if (time.Now().Unix() - t.Unix()) > 60 {
				continue
			}
			fmt.Printf("timestamp: %v, peer: %s, channel: %s, name: %s, value: %s\n", timestamp, peer, channel, name, val)
			topic := fmt.Sprintf("%s/%s/%s/%s", mqtt_topic_prefix, peer, channel, name)
			v, err := strconv.ParseUint(val, 16, 32)
			if err != nil {
				fmt.Printf("Conversion failed: %s\n", err)
				continue
			}
			var value float32
			if name == "IEC_ENERGY_COUNTER" {
				if v == 0 {
					fmt.Printf("Skipping possible invalid energy counter with value 0")
					continue
				}
				value = float32(int(v)) / 10000
			}
			if name == "IEC_POWER" {
				value = float32(int32(v)) / 100
			}
			publish(client, topic, value)
		}

	}

}
