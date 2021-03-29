package database

import (
	"database/sql"
	"errors"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"v2ray.com/core"
)

type DataBaseSubs struct {
	db        *sql.DB
	access    sync.RWMutex
	tableName string
}

func (obj *DataBaseSubs) BuildCommand(m map[string]interface{}) string {
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

func (dbj *DataBaseSubs) Get(m map[string]interface{}, limitList ...int) (*sql.Rows, error) {
	dbj.access.Lock()
	defer dbj.access.Unlock()
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
	command := "SELECT * FROM Subs" + dbj.BuildCommand(m) + " LIMIT " + strconv.Itoa(skip) + "," + strconv.Itoa(limit)
	log.Print(command)
	rows, err := dbj.db.Query(command)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
func (obj *DataBaseSubs) GetMap(m map[string]interface{}, limitList ...int) (map[string]interface{}, error) {
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
func (dbj *DataBaseSubs) Delete(m map[string]interface{}) error {
	if m["id"] == "" && m["name"] == "" && m["url"] == "" {
		return errors.New("config error")
	}
	dbj.access.Lock()
	defer dbj.access.Unlock()
	command := "DELETE FROM Subs " + dbj.BuildCommand(m)
	log.Debug(command)
	_, err := dbj.db.Exec(command)
	if err != nil {
		return err
	}
	return err
}

func (dbj *DataBaseSubs) Append(m map[string]interface{}) (int64, error) {
	if m["name"].(string) == "" || m["url"].(string) == "" {
		return -1, errors.New("config error")
	}
	dbj.access.Lock()
	defer dbj.access.Unlock()
	statment, err := dbj.db.Prepare(`
		INSERT INTO Subs(name, url) values(?,?)
	`)
	defer statment.Close()
	if err != nil {
		return -1, err
	}
	res, err := statment.Exec(m["name"], m["url"])
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, err
}

func (dbj *DataBaseSubs) Update(m map[string]interface{}) error {
	if m["name"].(string) == "" || m["url"].(string) == "" || m["id"].(string) == "" {
		return errors.New("config error")
	}
	dbj.access.Lock()
	defer dbj.access.Unlock()
	statment, err := dbj.db.Prepare(`
		UPDATE Subs SET name=?,url=? WHERE id=?
	`)
	defer statment.Close()
	if err != nil {
		return err
	}
	_, err = statment.Exec(m["name"].(string), m["url"].(string), m["id"].(string))
	if err != nil {
		return err
	}
	return nil
}

func (dbj *DataBaseSubs) ToConfig(m map[string]interface{}) (*core.Config, []string, error) {
	return nil, nil, nil
}

func (dbj *DataBaseSubs) ToMap(rows *sql.Rows) (map[string]interface{}, error) {
	var (
		id        int
		name, url string
	)
	if err := rows.Scan(&id, &name, &url); err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": id, "name": name, "url": url, "boundType": "sub"}, nil
}
func (dbj *DataBaseSubs) Tags(m map[string]interface{}) (map[string][]string, error) {
	return nil, nil
}
func (dbj *DataBaseSubs) Init(db *sql.DB) error {
	dbj.access.Lock()
	defer dbj.access.Unlock()
	dbj.tableName = "Subs"
	dbj.db = db
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS  Subs (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		url TEXT)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (dbj *DataBaseSubs) RegisterMap(m *map[string]DataBaseObj, db *sql.DB) {
	dbj.Init(db)
	(*m)["sub"] = dbj
}
