package app

import (
	"encoding/json"
)

//应用配置
func ApplyConfig(c map[string]json.RawMessage) {
	new(DataBaseController).ApplyConfig(c)
	new(StatusController).ApplyConfig(c)
	new(V2RayController).ApplyConfig(c)
	new(TemplateController).ApplyConfig(c)
}
