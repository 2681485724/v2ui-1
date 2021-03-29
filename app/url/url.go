package url

import "net/url"

type URLObj interface {
	ToMap(*url.URL) (map[string]interface{}, error)
}

var URLMap map[string]URLObj

func RegisterMap() (*map[string]URLObj, error) {
	URLMap = map[string]URLObj{}
	new(URLVmess).RegisterMap(&URLMap)
	new(URLss).RegisterMap(&URLMap)
	return &URLMap, nil
}
