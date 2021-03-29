package app

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

var Template *TemplateController

type TemplateController struct {
	templateMap map[string]string
}

func NewTemplateController() (*TemplateController, error) {
	if Template == nil {
		Template = &TemplateController{}
	}
	return Template, nil
}
func (t *TemplateController) ApplyConfig(c map[string]json.RawMessage) error {
	t, _ = NewTemplateController()
	err := json.Unmarshal(c["Templates"], &t.templateMap)
	if err != nil {
		return err
	}
	return nil
}

//注册路由,加载模板
func (t *TemplateController) RegisterRoutes(g *gin.Engine) {
	t, _ = NewTemplateController()
	g.LoadHTMLGlob("templates/*")
	for td, _ := range t.templateMap {
		g.GET(td, t.DealWithPath)
	}
}

func (t *TemplateController) DealWithPath(g *gin.Context) {
	g.HTML(http.StatusOK, t.templateMap[g.Request.URL.Path], gin.H{})
}
