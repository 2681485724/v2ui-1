package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhangguojvn/v2ui/app"
)

func loadConfig(path string) (map[string]json.RawMessage, error) {
	fileObj, err := os.Open(path)
	defer fileObj.Close()
	if err != nil {
		return nil, err
	}
	byteConfig, err := ioutil.ReadAll(fileObj)
	if err != nil {
		return nil, err
	}
	var config map[string]json.RawMessage
	json.Unmarshal(byteConfig, &config)
	return config, nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})
	config, _ := loadConfig("config.json")
	app.ApplyConfig(config) //必须在 app.RegisterRoutes 之前调用
	router := gin.New()
	router.Static("/assets", "./assets")
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	app.RegisterRoutes(router)
	var port string
	json.Unmarshal(config["PORT"], &port)
	router.Run(port)
}
