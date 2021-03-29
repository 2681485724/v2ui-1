package url

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

type URLVmess struct {
}

func (uv *URLVmess) Init() error {
	return nil
}
func (uv *URLVmess) ToMap(rowURL *url.URL) (map[string]interface{}, error) {
	link := rowURL.String()[8:]
	if i := len(link) % 4; i != 0 {
		link += strings.Repeat("=", 4-i)
	}
	data, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		return nil, err
	}
	jsonData := map[string]interface{}{}
	if json.Unmarshal(data, &jsonData) != nil {
		return nil, err
	}
	var streamSettings map[string]interface{}
	if jsonData["net"].(string) == "tcp" {
		hostList := strings.Split(jsonData["host"].(string), ",")
		streamSettings = map[string]interface{}{
			"network":  "tcp",
			"security": jsonData["tls"].(string),
			"tcpSettings": map[string]interface{}{
				"header": map[string]interface{}{
					"type": hostList,
				},
				"request": map[string]interface{}{
					"Host": jsonData["host"],
				},
			},
		}
	} else if jsonData["net"].(string) == "kcp" {
		streamSettings = map[string]interface{}{
			"network":  "kcp",
			"security": jsonData["tls"].(string),
			"kcpSettings": map[string]interface{}{
				"header": map[string]interface{}{
					"type": jsonData["type"].(string),
				},
			},
		}
	} else if jsonData["net"].(string) == "ws" {
		streamSettings = map[string]interface{}{
			"network":  "ws",
			"security": jsonData["tls"].(string),
			"wsSettings": map[string]interface{}{
				"path": jsonData["path"].(string),
				"headers": map[string]interface{}{
					"Host": jsonData["host"].(string),
				},
			},
		}
	} else if jsonData["net"].(string) == "h2" {
		hostList := strings.Split(jsonData["host"].(string), ",")
		streamSettings = map[string]interface{}{
			"network":  "h2",
			"security": jsonData["tls"].(string),
			"httpSettings": map[string]interface{}{
				"host": hostList,
				"path": jsonData["path"].(string),
			},
		}
	} else if jsonData["net"].(string) == "quic" {
		streamSettings = map[string]interface{}{
			"network":  "quic",
			"security": jsonData["tls"].(string),
			"quicSettings": map[string]interface{}{
				"security": jsonData["host"].(string),
				"key":      jsonData["path"].(string),
				"header": map[string]interface{}{
					"type": jsonData["type"].(string),
				},
			},
		}
	}

	config := map[string]interface{}{
		"name":         jsonData["ps"].(string),
		"group":        "default",
		"port":         jsonData["port"].(string),
		"boundType":    "outbound",
		"protocolType": "vmess",
		"protocolSettings": map[string]interface{}{
			"vnext": []interface{}{
				map[string]interface{}{
					"address": jsonData["add"].(string),
					"port":    jsonData["port"].(string),
					"users": []interface{}{
						map[string]interface{}{
							"id":       jsonData["id"].(string),
							"alterId":  jsonData["aid"].(string),
							"security": "auto",
						},
					},
				},
			},
		},
		"mux":            "0",
		"streamSettings": streamSettings,
		"proxySettings":  "null"}
	return config, nil
}
func (dbj *URLVmess) RegisterMap(m *map[string]URLObj) {
	dbj.Init()
	(*m)["vmess"] = dbj
}
