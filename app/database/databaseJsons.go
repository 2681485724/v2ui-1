package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"v2ray.com/core"
	jsonLoader "v2ray.com/core/infra/conf/serial"
)

type DataBaseJsons struct {
	db        *sql.DB
	access    sync.RWMutex
	tableName string
}

func (obj *DataBaseJsons) BuildCommand(m map[string]interface{}) string {
	if len(m) != 0 {
		var command string = " WHERE "
		for key, item := range m {
			command += key + "=\"" + item.(string) + "\","
		}
		command = command[:len(command)-1]
		return command
	}
	return ""
}
func (obj *DataBaseJsons) GetMap(m map[string]interface{}, limitList ...int) (map[string]interface{}, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if len(limitList) == 2 {
		rows, err = obj.Get(m, limitList[0], limitList[1])
	} else if len(limitList) == 1 {
		rows, err = obj.Get(m, limitList[0])
	} else {
		rows, err = obj.Get(m)
	}
	if err != nil {
		return nil, err
	}
	return obj.ToMap(rows)
}
func (obj *DataBaseJsons) Get(m map[string]interface{}, limitList ...int) (*sql.Rows, error) {
	obj.access.Lock()
	defer obj.access.Unlock()
	var skip, limit int
	if len(limitList) == 2 {
		limit = limitList[1]
		skip = limitList[0]
	} else if len(limitList) == 1 {
		limit = limitList[0]
		skip = 0
	} else {
		limit = 100
		skip = 0
	}
	command := "SELECT * FROM JSONs" + obj.BuildCommand(m) + " LIMIT " + strconv.Itoa(skip) + "," + strconv.Itoa(limit)
	log.Print(command)
	rows, err := obj.db.Query(command)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (obj *DataBaseJsons) Delete(m map[string]interface{}) error {
	if m["id"] == "" && m["name"] == "" && m["content"] == "" {
		return errors.New("config error")
	}
	obj.access.Lock()
	defer obj.access.Unlock()
	command := "DELETE FROM JSONs " + obj.BuildCommand(m)
	log.Debug(command)
	_, err := obj.db.Exec(command)
	if err != nil {
		return err
	}
	return err
}

func (obj *DataBaseJsons) Append(m map[string]interface{}) (int64, error) {
	if m["name"].(string) == "" || m["content"].(map[string]interface{}) == nil {
		return -1, errors.New("config error")
	}
	obj.access.Lock()
	defer obj.access.Unlock()
	content, err := json.Marshal(m["content"].(map[string]interface{}))
	if err != nil {
		return -1, err
	}
	statment, err := obj.db.Prepare(`
		INSERT INTO JSONs(name, content) values(?,?)
	`)
	defer statment.Close()
	if err != nil {
		return -1, err
	}
	res, err := statment.Exec(m["name"], content)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, err
}

func (obj *DataBaseJsons) Update(m map[string]interface{}) error {
	if m["name"].(string) == "" || m["content"].(map[string]interface{}) == nil || m["id"].(string) == "" {
		return errors.New("config error")
	}
	obj.access.Lock()
	defer obj.access.Unlock()
	content, err := json.Marshal(m["content"].(map[string]interface{}))
	statment, err := obj.db.Prepare(`
		UPDATE JSONs SET name=?,content=? WHERE id=?
	`)
	defer statment.Close()
	if err != nil {
		return err
	}
	_, err = statment.Exec(m["name"].(string), content, m["id"].(string))
	if err != nil {
		return err
	}
	return nil
}

func (obj *DataBaseJsons) ToConfig(m map[string]interface{}) (*core.Config, []string, error) {
	content, err := json.Marshal(m["content"])
	if err != nil {
		return nil, nil, err
	}
	configInput := bytes.NewBuffer(content)
	configObj, err := jsonLoader.LoadJSONConfig(configInput)
	if err != nil {
		return nil, nil, err
	}
	if configObj.Inbound != nil {
		for key, item := range configObj.Inbound {
			if item.Tag == "" {
				item.Tag = fmt.Sprintf("{'type':'json','id':'%d','number':'%d'}", m["id"], key)
			}
		}
	}
	if configObj.Outbound != nil {
		for key, item := range configObj.Outbound {
			if item.Tag == "" {
				item.Tag = fmt.Sprintf("{'type':'json','id':'%d','number':'%d'}", m["id"], key)
			}
		}
	}
	return configObj, []string{m["id"].(string)}, nil
}

func (obj *DataBaseJsons) ToMap(rows *sql.Rows) (map[string]interface{}, error) {
	var (
		id            int
		name, content string
		c             map[string]interface{}
	)
	if err := rows.Scan(&id, &name, &content); err != nil {
		return nil, err
	}
	var result map[string]interface{}

	err := json.Unmarshal([]byte(content), &c)
	if err != nil {
		return nil, err
	}
	result["content"] = c
	result["id"] = id
	result["name"] = name
	result["boundType"] = "json"
	return result, nil
}
func (obj *DataBaseJsons) Tags(m map[string]interface{}) (map[string][]string, error) {
	rows, err := obj.Get(m)
	if err != nil {
		return nil, err
	}
	rows.Next()
	configMap, err := obj.ToMap(rows)
	if err != nil {
		return nil, err
	}
	configObj, _, err := obj.ToConfig(configMap)
	if err != nil {
		return nil, err
	}
	tag := map[string][]string{"inbounds": nil, "outbounds": nil}
	if configObj.Inbound != nil {
		for _, item := range configObj.Inbound {
			tag["inbounds"] = append(tag["inbounds"], item.Tag)
		}
	}
	if configObj.Outbound != nil {
		for _, item := range configObj.Outbound {
			tag["outbounds"] = append(tag["outbounds"], item.Tag)
		}

	}
	return tag, nil
}
func (obj *DataBaseJsons) Init(db *sql.DB) error {
	obj.access.Lock()
	defer obj.access.Unlock()
	obj.tableName = "JSONs"
	obj.db = db
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS  JSONs (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		content TEXT)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (obj *DataBaseJsons) RegisterMap(m *map[string]DataBaseObj, db *sql.DB) {
	obj.Init(db)
	(*m)["json"] = obj
}
