package url

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type URLss struct {
}

func (uv *URLss) Init() error {
	return nil
}
func (uv *URLss) ToMap(rowURL *url.URL) (map[string]interface{}, error) {
	host := strings.Split(rowURL.Host, ":")
	rawUserInfo := rowURL.User.String()
	if i := len(rawUserInfo) % 4; i != 0 {
		rawUserInfo += strings.Repeat("=", 4-i)
	}
	rowUserinfo, err := base64.StdEncoding.DecodeString(rawUserInfo)
	if err != nil {
		return nil, errors.New(rowURL.User.String())
	}
	userinfo := strings.Split(string(rowUserinfo), ":")
	if rowURL.Fragment == "" {
		rowURL.Fragment = "default"
	}
	port, err := strconv.ParseInt(rowURL.Port(), 10, 64)
	if err != nil {
		return nil, err
	}
	config := map[string]interface{}{
		"name":         rowURL.Fragment,
		"group":        "default",
		"port":         host[1],
		"boundType":    "outbound",
		"protocolType": "shadowsocks",
		"protocolSettings": map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"address":  rowURL.Hostname(),
					"port":     port,
					"method":   userinfo[0],
					"password": userinfo[1],
				},
			},
		},
		"mux":            "0",
		"streamSettings": nil,
		"proxySettings":  nil}
	return config, nil
}
func (dbj *URLss) RegisterMap(m *map[string]URLObj) {
	dbj.Init()
	(*m)["ss"] = dbj
}
