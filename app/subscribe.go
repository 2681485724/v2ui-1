package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	urlLoader "net/url"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SUBController struct {
}

//添加订阅
func (sub *SUBController) append(jsonData map[string]interface{}) (int64, error) {
	id, err := (*DataBaseMap)["sub"].Append(map[string]interface{}{"name": jsonData["name"], "url": jsonData["url"]})
	if err != nil {
		return -1, err
	}
	return id, nil
}

//查找满足条件的订阅
func (sub *SUBController) get(m *map[string]interface{}) ([]map[string]interface{}, error) {
	rows, err := (*DataBaseMap)["sub"].Get(*m)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var result = []map[string]interface{}{}
	for rows.Next() {
		configMap, err := (*DataBaseMap)["sub"].ToMap(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, configMap)
	}
	return result, nil
}

//删除指定的订阅,并返回其内容
func (sub *SUBController) pop(m *map[string]interface{}) ([]map[string]interface{}, error) {
	result, err := sub.get(m)
	if err != nil {
		return nil, err
	}
	if err := (*DataBaseMap)["sub"].Delete(*m); err != nil {
		return nil, err
	}
	return result, nil
}

//删除指定订阅,及其配置文件
func (sub *SUBController) remove(m *map[string]interface{}) error {
	rawdata, err := sub.pop(m)
	if err != nil {
		return err
	}
	for _, item := range rawdata {
		if err := (*DataBaseMap)["formatted"].Delete(map[string]interface{}{"fgroup": fmt.Sprintf("sub-%d", item["id"].(int))}); err != nil {
			return err
		}
	}
	return nil
}

//更新订阅内容
func (sub *SUBController) update(m *[]map[string]interface{}) error {
	for _, item := range *m {
		client := &http.Client{}

		req, err := http.NewRequest("GET", item["url"].(string), nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/74.0")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		rowURL, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return err
		}
		URLs := strings.Split(string(rowURL), "\n")

		for _, url := range URLs {
			url = strings.Replace(url, "\r", "", -1)
			url = strings.Replace(url, "\t", "", -1)
			if url == "" {
				continue
			}
			log.Info(item)
			rowURL, err := urlLoader.Parse(url)
			if err != nil {
				continue
			}
			configMap, err := (*URLMap)[rowURL.Scheme].ToMap(rowURL)
			if err != nil {
				continue
			}
			configMap["group"] = fmt.Sprintf("sub-%d", item["id"].(int))
			_, err = (*DataBaseMap)["formatted"].Append(configMap)
			if err != nil {
				continue
			}
		}
	}
	return nil
}

func (sub *SUBController) Append(c *gin.Context) {
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	_, err := sub.append(jsonData)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	c.JSON(200, nil)
}

func (sub *SUBController) List(c *gin.Context) {
	rawData, err := sub.get(new(map[string]interface{}))
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	var result []map[string]interface{}
	for _, item := range rawData {
		result = append(result, map[string]interface{}{"name": item["name"],
			"id":        map[string]interface{}{"id": item["id"], "type": "sub"},
			"boundType": item["boundType"]})

	}
	c.JSON(200, result)
}

func (sub *SUBController) Get(c *gin.Context) {
	rawData, err := sub.get(&map[string]interface{}{"id": c.Param("id")})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	c.JSON(200, rawData[0])
}

func (sub *SUBController) Cmd(c *gin.Context) {
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	if "del" == jsonData["cmd"].(string) {
		err := sub.remove(&map[string]interface{}{"id": c.Param("id")})
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
	} else if "edit" == jsonData["cmd"].(string) {
		jsonData["id"] = c.Param("id")
		if err := (*DataBaseMap)["sub"].Update(jsonData); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		c.JSON(200, nil)
	} else if "update" == jsonData["cmd"].(string) {
		rawData, err := sub.get(&map[string]interface{}{"id": c.Param("id")})
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		if rawData == nil {
			c.JSON(404, gin.H{"error": "can't find"})
			return
		}
		if err := (*DataBaseMap)["formatted"].Delete(map[string]interface{}{"fgroup": fmt.Sprintf("sub-%d", rawData[0]["id"].(int))}); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		if err := sub.update(&rawData); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		c.JSON(200, nil)

	}
}

func (sub *SUBController) RegisterRoutes(g *gin.Engine) {
	g.GET("/api/sub", sub.List)
	g.POST("/api/sub", sub.Append)
	g.GET("/api/sub/:id", sub.Get)
	g.POST("/api/sub/:id", sub.Cmd)
}
