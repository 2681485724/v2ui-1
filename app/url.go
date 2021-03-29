package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	urlLoader "net/url"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhangguojvn/v2ui/app/url"
)

var (
	URLMap *map[string]url.URLObj
	URL    *URLController
)

type URLController struct {
}

func NewURLController() (*URLController, error) {
	if URL == nil {
		URL = new(URLController)
		var err error
		URLMap, err = url.RegisterMap()
		return nil, err
	}
	return URL, nil
}
func (c *URLController) addURLs(g *gin.Context) {
	var jsonData map[string]interface{}
	data, _ := ioutil.ReadAll(g.Request.Body)
	if err := json.Unmarshal(data, &jsonData); err != nil {
		g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
		log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
		return
	}
	for _, item := range jsonData["URLs"].([]interface{}) {
		rowURL, err := urlLoader.Parse(item.(string))
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		configMap, err := (*URLMap)[rowURL.Scheme].ToMap(rowURL)
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
		configMap["group"] = jsonData["group"].(string)
		_, err = (*DataBaseMap)["formatted"].Append(configMap)
		if err != nil {
			g.JSON(500, gin.H{"error": fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error())})
			log.Warn(fmt.Sprintf("%s:%s", reflect.TypeOf(err).PkgPath(), err.Error()))
			return
		}
	}
}
func (c *URLController) RegisterRoutes(g *gin.Engine) {
	c, err := NewURLController()
	if err != nil {
		log.Error(err)
	}
	g.POST("/api/url", c.addURLs)
}
