package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
}

func (c *ConfigController) append(m map[string]interface{}) error {
	_, err := (*DataBaseMap)[m["type"].(string)].Append(m)
	return err
}

func (c *ConfigController) Append(g *gin.Context) {
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(g.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		g.JSON(500, gin.H{"error": fmt.Sprintf(err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	jsonData["type"] = g.Param("type")
	err := c.append(jsonData)
	if err != nil {
		g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	g.JSON(200, nil)
}
func (c *ConfigController) searchConfigs(g *gin.Context) {
	result := []map[string]interface{}{}
	for datebaseType, function := range *DataBaseMap {
		rows, err := function.Get(map[string]interface{}{})
		defer rows.Close()
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		for rows.Next() {
			configMap, err := function.ToMap(rows)
			if err != nil {
				g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
				log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
				return
			}
			result = append(result, map[string]interface{}{"name": configMap["name"],
				"id":        map[string]interface{}{"id": configMap["id"], "type": datebaseType},
				"boundType": configMap["boundType"]})
		}
	}
	g.JSON(200, result)
}
func (c *ConfigController) getConfigs(g *gin.Context) {
	rows, err := (*DataBaseMap)[g.Param("type")].Get(map[string]interface{}{"id": g.Param("id")})
	defer rows.Close()
	if err != nil {
		g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	if rows.Next() {
		result, err := (*DataBaseMap)[g.Param("type")].ToMap(rows)
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		g.JSON(200, result)
	} else {
		g.JSON(404, nil)
	}

}
func (c *ConfigController) update(m map[string]interface{}) error {
	err := (*DataBaseMap)[m["type"].(string)].Update(m)
	return err
}
func (c *ConfigController) editConfig(g *gin.Context) {
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(g.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	if "del" == jsonData["cmd"].(string) {
		if err := (*DataBaseMap)[g.Param("type")].Delete(map[string]interface{}{"id": g.Param("id")}); err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		g.JSON(200, nil)
	} else if "edit" == jsonData["cmd"].(string) {
		jsonData["id"] = g.Param("id")
		jsonData["type"] = "formatted"
		err := c.update(jsonData)
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		g.JSON(200, nil)
	}
}
func (c *ConfigController) RegisterRoutes(g *gin.Engine) {
	g.POST("/api/configs/:type", c.Append)
	g.GET("/api/configs", c.searchConfigs)
	g.GET("/api/configs/:type/:id", c.getConfigs)
	g.POST("/api/configs/:type/:id", c.editConfig)
}
