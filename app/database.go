package app

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zhangguojvn/v2ui/app/database"
)

var (
	DataBase    *DataBaseController
	DataBaseMap *map[string]database.DataBaseObj
)

type DataBaseController struct {
	db *sql.DB
}

//初始化数据库,单例.
func NewDataBaseController(path string) (*DataBaseController, error) {
	if DataBase == nil {
		d := &DataBaseController{}
		DataBase = d
		var err error
		d.db, err = sql.Open("sqlite3", path)
		if err != nil {
			return nil, errors.New("failed to init data base")
		}
		DataBaseMap, err = database.RegisterMap(d.db)
		if err != nil {
			return nil, err
		}
	}
	return DataBase, nil
}

//应用配置
func (d *DataBaseController) ApplyConfig(c map[string]json.RawMessage) error {
	var path string
	var err error
	err = json.Unmarshal(c["DataBasePath"], &path)
	if err != nil {
		return err
	}
	d, err = NewDataBaseController(path)
	if err != nil {
		return err
	}

	return nil
}

func (self DataBaseController) RegisterRoutes(g *gin.Engine) {
}
