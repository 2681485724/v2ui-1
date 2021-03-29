package app_test

import (
	"bytes"
	testNet "net"
	"testing"

	. "github.com/zhangguojvn/v2ui/app"
	"v2ray.com/core/common"
	jsonLoader "v2ray.com/core/infra/conf/serial"
	_ "v2ray.com/core/main/distro/all"
	"v2ray.com/core/testing/servers/tcp"
)

func TestV2Ray(t *testing.T) {
	port := tcp.PickPort()
	portString := port.String()
	config := `
	{
		"inbounds": [
			{
				"port": ` + portString + `,
				"listen": "127.0.0.1",
				"protocol": "socks",
				"settings": {},
				"streamSettings": {},
				"tag": "def",
				"sniffing": {
				  "enabled": false,
				  "destOverride": ["http", "tls"]
				},
				"allocate": {
				  "strategy": "always",
				  "refresh": 5,
				  "concurrency": 3
				}
			  }
		],
		"outbounds": [
		  	{
				  "tag":"test",
				"protocol": "freedom",  
				"settings": {}
			  }
			]
		  }
	`
	server, err := NewV2RayController()
	common.Must(err)
	err = server.LoadJSONConfig(config)
	common.Must(err)
	err = server.Start()
	common.Must(err)
	conn, err := testNet.Dial("tcp", ":"+portString)
	common.Must(err)
	conn.Close()
	server.Stop()
	_, err = testNet.Dial("tcp", ":"+portString)
	if err == nil {
		panic("cann't stop")
	}
	server.Start()
	conn, err = testNet.Dial("tcp", ":"+portString)
	common.Must(err)
	conn.Close()
	err = server.RemoveInboundConfig("def")
	common.Must(err)
	_, err = testNet.Dial("tcp", ":"+portString)
	if err == nil {
		panic("cann't remove")
	}
	configInput := bytes.NewBuffer([]byte(config))
	configObj, err := jsonLoader.LoadJSONConfig(configInput)
	if err != nil {
		panic(err)
	}
	server.AddInboundConfig(configObj.Inbound[0])
	server.AddOutboundConfig(configObj.Outbound[0])
	conn, err = testNet.Dial("tcp", ":"+portString)
	common.Must(err)
	conn.Close()
	server.RemoveOutboundConfig("test")
	server.Stop()
}

func TestDatabase(t *testing.T) {

}
