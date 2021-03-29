package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"v2ray.com/core"
	"v2ray.com/core/app/dispatcher"
	"v2ray.com/core/app/proxyman"
	inboundManager "v2ray.com/core/app/proxyman/inbound"
	outboundManager "v2ray.com/core/app/proxyman/outbound"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/features/inbound"
	"v2ray.com/core/features/outbound"
	jsonLoader "v2ray.com/core/infra/conf/serial"
)

var V2rayID struct {
}
var V2Ray *V2RayController

type V2RayController struct {
	access          sync.RWMutex
	core            *core.Instance
	counter         int
	outboundManager *outboundManager.Manager
	outboundList    []string
	inboundManager  *inboundManager.Manager
	inboundList     []string
	running         bool
	config          *core.Config
}

//初始化内核
func NewV2RayController() (*V2RayController, error) {
	if V2Ray == nil {
		v := &V2RayController{}
		V2Ray = v
		var err error
		v.config = &core.Config{
			App: []*serial.TypedMessage{
				serial.ToTypedMessage(&dispatcher.Config{}),
				serial.ToTypedMessage(&proxyman.InboundConfig{}),
				serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			}}
		v.core, err = core.New(v.config)
		if err != nil {
			return nil, err
		}
		v.core.Start() //开启内核
		v.running = true

		v.outboundManager, _ = v.core.GetFeature(outbound.ManagerType()).(*outboundManager.Manager)
		v.inboundManager, _ = v.core.GetFeature(inbound.ManagerType()).(*inboundManager.Manager)
	}
	return V2Ray, nil
}

//启动
func (v *V2RayController) Start() error {
	if !(v.running) {
		v.access.Lock()
		defer v.access.Unlock()
		v.running = true
		if err := v.inboundManager.Start(); err != nil {
			return err
		}
		if err := v.outboundManager.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (v *V2RayController) Status() bool {
	return v.running
}

func (v *V2RayController) Stop() error {
	if v.running {
		v.access.Lock()
		defer v.access.Unlock()
		v.running = false
		if err := v.inboundManager.Close(); err != nil {
			return err
		}
		if err := v.outboundManager.Close(); err != nil {
			return err
		}
	}
	return nil
}

//添加出站协议
func (v *V2RayController) AddOutboundConfig(config *core.OutboundHandlerConfig) error {
	v.access.Lock()
	defer v.access.Unlock()
	if config.Tag == "" {
		v.counter++
		config.Tag = fmt.Sprintf("untaggedHandler_%s", strconv.Itoa(v.counter))
	}
	v.outboundList = append(v.outboundList, config.Tag)
	rawHandler, err := core.CreateObject(v.core, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return errors.New("not an OutboundHandler")
	}
	v.outboundManager.AddHandler(context.Background(), handler)
	return nil
}

//批量添加出站协议
func (v *V2RayController) AddOutboundConfigs(config []*core.OutboundHandlerConfig) error {
	var err error
	for _, outboundConfig := range config {
		err = v.AddOutboundConfig(outboundConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
func (v *V2RayController) RemoveOutboundConfig(tag string) error {
	handler := (*v.outboundManager).GetHandler(tag)
	err := handler.Close()
	if err != nil {
		return err
	}
	return (*v.outboundManager).RemoveHandler(context.Background(), tag)

}

//添加进站协议
func (v *V2RayController) AddInboundConfig(config *core.InboundHandlerConfig) error {
	v.access.Lock()
	defer v.access.Unlock()
	if config.Tag == "" {
		v.counter++
		config.Tag = fmt.Sprintf("untaggedHandler_%s", strconv.Itoa(v.counter))
	}
	v.inboundList = append(v.inboundList, config.Tag)
	rawHandler, err := core.CreateObject(v.core, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(inbound.Handler)
	if !ok {
		return errors.New("not an InboundHandler")
	}
	v.inboundManager.AddHandler(context.Background(), handler)
	return nil
}

//批量添加进站协议
func (v *V2RayController) AddInboundConfigs(config []*core.InboundHandlerConfig) error {
	var err error
	for _, inboundConfig := range config {
		err = v.AddInboundConfig(inboundConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *V2RayController) RemoveInboundConfig(tag string) error {
	return v.inboundManager.RemoveHandler(context.Background(), tag)

}

//读取JSON数据
func (v *V2RayController) LoadJSONConfig(config string) error {
	configInput := bytes.NewBuffer([]byte(config))
	configObj, err := jsonLoader.LoadJSONConfig(configInput)
	if err != nil {
		return err
	}
	v.AddInboundConfigs(configObj.Inbound)
	v.AddOutboundConfigs(configObj.Outbound)
	return nil
}

func (v *V2RayController) ClearOutbound() error {
	v.access.Lock()
	defer v.access.Unlock()
	for _, tag := range v.outboundList {
		if err := v.RemoveOutboundConfig(tag); err != nil {
			return err
		}
	}
	return nil
}

func (v *V2RayController) ClearInbound() error {
	v.access.Lock()
	defer v.access.Unlock()
	for _, tag := range v.inboundList {
		if err := v.RemoveInboundConfig(tag); err != nil {
			return err
		}
	}
	return nil
}
func (v *V2RayController) GetOutList() []string {
	return v.outboundList
}
func (v *V2RayController) GetInList() []string {
	return v.inboundList
}

//清除所有配置
func (v *V2RayController) Clear() error {
	if err := v.ClearInbound(); err != nil {
		return err
	}
	if err := v.ClearOutbound(); err != nil {
		return err
	}
	return nil
}
func (v *V2RayController) ApplyConfig(c map[string]json.RawMessage) error {
	v.access.Lock()
	defer v.access.Unlock()
	v, _ = NewV2RayController()
	return nil
}
func (v *V2RayController) RegisterRoutes(g *gin.Engine) {
	v.access.Lock()
	defer v.access.Unlock()
	v, _ = NewV2RayController()
}
