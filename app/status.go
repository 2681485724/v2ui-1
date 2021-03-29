package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type StatusController struct {
	access   sync.RWMutex
	configId map[string]map[string][]string
}

func (s *StatusController) ApplyConfig(c map[string]json.RawMessage) error {
	return nil
}
func (self *StatusController) RegisterRoutes(g *gin.Engine) {
	g.POST("/api/status", self.ChangeStatus)
	g.GET("/api/status", self.GetStatus)
	g.POST("/test", self.Test)
}
func (self *StatusController) GetStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  V2Ray.Status(),
		"running": self.configId,
	})
}
func (self *StatusController) ChangeStatus(c *gin.Context) {
	self.access.Lock()
	defer self.access.Unlock()
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	if jsonData["cmd"].(string) == "add" {

		config, err := (*DataBaseMap)[jsonData["id"].(map[string]interface{})["type"].(string)].Get(
			map[string]interface{}{"id": jsonData["id"].(map[string]interface{})["id"].(string)})
		defer config.Close()
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		config.Next()
		configMap, err := (*DataBaseMap)[jsonData["id"].(map[string]interface{})["type"].(string)].ToMap(config)
		config.Close()
		configObj, ids, err := (*DataBaseMap)[jsonData["id"].(map[string]interface{})["type"].(string)].ToConfig(configMap)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		if configObj.Inbound != nil {
			V2Ray.AddInboundConfigs(configObj.Inbound)
		}
		if configObj.Outbound != nil {
			V2Ray.AddOutboundConfigs(configObj.Outbound)
		}
		V2Ray.Start()
		for _, id := range ids {
			tag, err := (*DataBaseMap)[jsonData["id"].(map[string]interface{})["type"].(string)].
				Tags(map[string]interface{}{"id": id})
			if err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
				log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
				return
			}
			if self.configId == nil {
				self.configId = map[string]map[string][]string{}
			}
			self.configId[fmt.Sprintf("{'type':'%s','id':'%s'}",
				jsonData["id"].(map[string]interface{})["type"].(string),
				id)] = tag
		}
		c.JSON(200, nil)
	} else if jsonData["cmd"].(string) == "stop" {
		err := V2Ray.Stop()
		if err != nil {
			log.Print(err)
		}
		c.JSON(200, nil)
	} else if jsonData["cmd"].(string) == "restart" {
		V2Ray.Stop()
		V2Ray.Start()
		c.JSON(200, nil)
	} else if jsonData["cmd"].(string) == "clear" {
		err := V2Ray.Clear()
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Print(err)
		}
	} else if jsonData["cmd"].(string) == "del" {
		tag := self.configId[fmt.Sprintf("{'type':'%s','id':'%s'}",
			jsonData["id"].(map[string]interface{})["type"].(string),
			jsonData["id"].(map[string]interface{})["id"].(string))]
		if tag["outbounds"] != nil {
			for _, item := range tag["outbounds"] {
				if err := V2Ray.RemoveOutboundConfig(item); err != nil {
					c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
					log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
					return
				}
			}
		}
		if tag["inbounds"] != nil {
			for _, item := range tag["inbounds"] {
				if err := V2Ray.RemoveInboundConfig(item); err != nil {
					c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
					log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
					return
				}
			}
		}
		delete(self.configId, fmt.Sprintf("{'type':'%s','id':'%s'}",
			jsonData["id"].(map[string]interface{})["type"].(string),
			jsonData["id"].(map[string]interface{})["id"].(string)))
		c.JSON(200, nil)
	}
}
func (self *StatusController) Test(c *gin.Context) {
	content := c.PostForm("input")
	V2Ray.LoadJSONConfig(content)
	V2Ray.Start()
	c.JSON(200, gin.H{
		"status": V2Ray.Status(),
	})
}
