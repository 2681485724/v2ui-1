package database

import (
	"database/sql"

	"v2ray.com/core"
)

type DataBaseObj interface {
	Get(map[string]interface{}, ...int) (*sql.Rows, error)
	GetMap(map[string]interface{}, ...int) (map[string]interface{}, error)
	Delete(map[string]interface{}) error
	Append(map[string]interface{}) (int64, error)
	Update(map[string]interface{}) error
	ToConfig(map[string]interface{}) (*core.Config, []string, error)
	ToMap(*sql.Rows) (map[string]interface{}, error)
	Tags(map[string]interface{}) (map[string][]string, error)
}

var DataBaseMap map[string]DataBaseObj

func RegisterMap(db *sql.DB) (*map[string]DataBaseObj, error) {
	DataBaseMap = map[string]DataBaseObj{}
	new(DataBaseJsons).RegisterMap(&DataBaseMap, db)
	new(DataBaseFormatteds).RegisterMap(&DataBaseMap, db)
	new(DataBaseSubs).RegisterMap(&DataBaseMap, db)
	return &DataBaseMap, nil
}
