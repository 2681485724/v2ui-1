package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"v2ray.com/core/infra/conf"

	"v2ray.com/core"
)

type DataBaseFormatteds struct {
	db        *sql.DB
	access    sync.RWMutex
	tableName string
}

func (obj *DataBaseFormatteds) BuildCommand(m map[string]interface{}) string {
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
func (obj *DataBaseFormatteds) isCycle(thisID string, lastID string) bool {
	var configMap map[string]interface{}
	for {
		rows, err := obj.Get(map[string]interface{}{"id": lastID})
		defer rows.Close()
		if err != nil {
			return true
		}
		if rows.Next() {
			configMap, err = obj.ToMap(rows)
			if err != nil {
				return true
			}
		} else {
			return true
		}
		rows.Close()
		if configMap["id"] == thisID {
			return true
		} else {
			if configMap["proxyID"] == "" {
				break
			} else {
				lastID = configMap["proxyID"].(string)
			}
		}
	}
	return false
}
func (obj *DataBaseFormatteds) GetMap(m map[string]interface{}, limitList ...int) (map[string]interface{}, error) {
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
func (dbj *DataBaseFormatteds) Get(m map[string]interface{}, limitList ...int) (*sql.Rows, error) {
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
	command := "SELECT * FROM Formatteds" + dbj.BuildCommand(m) + " LIMIT " + strconv.Itoa(skip) + "," + strconv.Itoa(limit)
	log.Print(command)
	rows, err := dbj.db.Query(command)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (dbj *DataBaseFormatteds) Delete(m map[string]interface{}) error {
	for _, item := range m {
		if item != "" {
			goto a
		}
	}
	return errors.New("config error")

a:
	dbj.access.Lock()
	defer dbj.access.Unlock()
	command := "DELETE FROM Formatteds " + dbj.BuildCommand(m)
	log.Print(command)
	_, err := dbj.db.Exec(command)
	if err != nil {
		return err
	}
	return err
}

func (dbj *DataBaseFormatteds) Append(m map[string]interface{}) (int64, error) {
	if m["name"].(string) == "" ||
		m["group"].(string) == "" ||
		m["boundType"].(string) == "" ||
		m["protocolType"].(string) == "" {
		return -1, errors.New("config error")
	}
	dbj.access.Lock()
	defer dbj.access.Unlock()
	protocolSettings, err := json.Marshal(m["protocolSettings"])
	if err != nil {
		return -1, err
	}
	streamSettings, err := json.Marshal(m["streamSettings"])
	if err != nil {
		return -1, err
	}
	statment, err := dbj.db.Prepare(`
		INSERT INTO Formatteds( 
			name, 
			fgroup,
			port,
			boundType,
			protocolType,
			protocolSettings,
			mux,
			streamSettings,
			proxyID) values(?,?,?,?,?,?,?,?,?)
	`)
	defer statment.Close()
	if err != nil {
		return -1, err
	}
	res, err := statment.Exec(
		m["name"].(string),
		m["group"].(string),
		m["port"],
		m["boundType"].(string),
		m["protocolType"].(string),
		protocolSettings,
		m["mux"].(string),
		streamSettings,
		m["proxyID"],
	)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, err
}

func (dbj *DataBaseFormatteds) Update(m map[string]interface{}) error {
	if m["name"].(string) == "" ||
		m["group"].(string) == "" ||
		m["boundType"].(string) == "" ||
		m["protocolType"].(string) == "" ||
		m["id"].(string) == "" {
		return errors.New("config error")
	}
	if m["proxyID"] != nil && dbj.isCycle(m["id"].(string), m["proxyID"].(string)) {
		return errors.New("Cycle")
	}
	dbj.access.Lock()
	defer dbj.access.Unlock()
	protocolSettings, err := json.Marshal(m["protocolSettings"])
	if err != nil {
		return err
	}
	streamSettings, err := json.Marshal(m["streamSettings"])
	if err != nil {
		return err
	}

	statment, err := dbj.db.Prepare(`
		UPDATE Formatteds SET 
		name = ?, 
		fgroup = ?,
		port = ?,
		boundType = ?,
		protocolType = ?,
		protocolSettings = ?,
		mux = ?,
		streamSettings = ?,
		proxyID = ? WHERE id=?
	`)
	defer statment.Close()
	if err != nil {
		return err
	}
	_, err = statment.Exec(
		m["name"],
		m["group"],
		m["port"],
		m["boundType"],
		m["protocolType"],
		protocolSettings,
		m["mux"],
		streamSettings,
		m["proxyID"],
		m["id"])
	if err != nil {
		return err
	}
	return nil
}

func (dbj *DataBaseFormatteds) ToOutbound(m map[string]interface{}) (conf.OutboundDetourConfig, error) {

	outboundObj := conf.OutboundDetourConfig{}
	protocolSettings, err := json.Marshal(m["protocolSettings"])
	if err != nil {
		return outboundObj, nil
	}
	streamSettings, err := json.Marshal(m["streamSettings"])
	if err != nil {
		return outboundObj, nil
	}
	proxySettings, err := json.Marshal(m["proxySettings"])
	if err != nil {
		return outboundObj, nil
	}
	input := []byte(fmt.Sprintf(`
		{
			"sendThrough": "0.0.0.0",
			"protocol": "%s",
			"settings": %s,
			"tag": "%s",
			"streamSettings": %s,
			"proxySettings": %s,
			"mux": {
				"enable": %t,
				"concurrency": %s
			}
		}
		`, m["protocolType"].(string),
		protocolSettings,
		fmt.Sprintf(`{'type':'formatted','id':'%d'}`,
			m["id"].(string)),
		streamSettings,
		proxySettings,
		m["mux"].(string) != "0",
		m["mux"].(string),
	))
	err = json.Unmarshal(input, &outboundObj)
	if err != nil {
		return outboundObj, err
	}
	return outboundObj, nil
}

func (dbj *DataBaseFormatteds) ToInbound(m map[string]interface{}) (conf.InboundDetourConfig, error) {
	inboundObj := conf.InboundDetourConfig{}
	protocolSettings, err := json.Marshal(m["protocolSettings"])
	if err != nil {
		return inboundObj, nil
	}
	streamSettings, err := json.Marshal(m["streamSettings"])
	if err != nil {
		return inboundObj, nil
	}
	input := []byte(fmt.Sprintf(`
	{
		"port": "%s",
		  "listen": "0.0.0.0",
		  "protocol": "%s",
		  "settings": %s,
		  "streamSettings": %s,
		  "tag": "%s",
		  "sniffing": {
			"enabled": false,
			"destOverride": ["http", "tls"]
		  },
		  "allocate": {
		"strategy": "always",
		"refresh": 5,
		"concurrency": 3
		}
	  }
	`, m["port"].(string),
		m["protocolType"].(string),
		protocolSettings,
		streamSettings,
		fmt.Sprintf(`{'type':'formatted','id':'%d'}`,
			m["id"].(string)),
	))
	err = json.Unmarshal(input, &inboundObj)
	if err != nil {
		return inboundObj, err
	}
	return inboundObj, nil
}

func (dbj *DataBaseFormatteds) ToConfig(m map[string]interface{}) (*core.Config, []string, error) {
	var ids []string = []string{}
	config := new(conf.Config)
	configMap := m
	if configMap["boundType"] == "inbound" {
		inboundObj, err := dbj.ToInbound(configMap)
		if err != nil {
			return nil, nil, err
		}
		config.InboundConfigs = []conf.InboundDetourConfig{inboundObj}
		ids = append(ids, configMap["id"].(string))
	} else if configMap["boundType"] == "outbound" {
		config.OutboundConfigs = []conf.OutboundDetourConfig{}
		for {
			outboundObj, err := dbj.ToOutbound(configMap)
			if err != nil {
				return nil, nil, err
			}
			config.OutboundConfigs = append(config.OutboundConfigs, outboundObj)
			ids = append(ids, configMap["id"].(string))
			if configMap["proxyID"] == "" {
				break
			} else {
				rows, err := dbj.Get(map[string]interface{}{"id": configMap["proxyID"]})
				defer rows.Close()
				if err != nil {
					return nil, nil, err
				}
				if rows.Next() {
					configMap, err = dbj.ToMap(rows)
					if err != nil {
						return nil, nil, err
					}
				} else {
					return nil, nil, errors.New("")
				}
				rows.Close()
			}
		}
	}
	configObj, err := config.Build()
	if err != nil {
		return nil, nil, err
	}
	return configObj, ids, nil
}

func (dbj *DataBaseFormatteds) ToMap(rows *sql.Rows) (map[string]interface{}, error) {
	var (
		id, port, mux, proxyID                                                 sql.NullInt64
		name, group, boundType, protocolType, protocolSettings, streamSettings sql.NullString
		result                                                                 map[string]interface{} = map[string]interface{}{}
	)
	if err := rows.Scan(&id, &name, &group, &port, &boundType, &protocolType, &protocolSettings, &mux, &streamSettings, &proxyID); err != nil {
		return nil, err
	}
	if id.Valid {
		result["id"] = strconv.FormatInt(id.Int64, 10)
	} else {
		return nil, nil
	}
	if name.Valid {
		result["name"] = name.String
	} else {
		return nil, nil
	}
	if group.Valid {
		result["group"] = group.String
	} else {
		result["group"] = ""
	}
	if port.Valid {
		result["port"] = strconv.FormatInt(port.Int64, 10)
	} else {
		result["port"] = int64(0)
	}
	if boundType.Valid {
		result["boundType"] = boundType.String
	} else {
		return nil, nil
	}
	if protocolType.Valid {
		result["protocolType"] = protocolType.String
	} else {
		return nil, nil
	}
	if mux.Valid {
		result["mux"] = strconv.FormatInt(mux.Int64, 10)
	} else {
		result["mux"] = "0"
	}
	if streamSettings.Valid {
		stream := map[string]interface{}{}
		err := json.Unmarshal([]byte(streamSettings.String), &stream)
		if err != nil {
			return nil, err
		}
		result["streamSettings"] = stream
	} else {
		result["streamSettings"] = map[string]interface{}{}
	}
	if protocolSettings.Valid {
		protocol := map[string]interface{}{}
		err := json.Unmarshal([]byte(protocolSettings.String), &protocol)
		if err != nil {
			return nil, err
		}
		result["protocolSettings"] = protocol
	} else {
		result["protocolSettings"] = map[string]interface{}{}
	}
	if proxyID.Valid {

		result["proxyID"] = strconv.FormatInt(proxyID.Int64, 10)
		tags, err := dbj.Tags(map[string]interface{}{"id": result["proxyID"]})
		if err != nil {
			return nil, err
		}
		if tags == nil || tags["outbounds"] == nil {
			result["proxyID"] = ""
		} else {
			result["proxySettings"] = map[string]interface{}{
				"tag": tags["outbounds"][0]}
		}
	} else {
		result["proxyID"] = ""
	}
	return result, nil
}

func (dbj *DataBaseFormatteds) Init(db *sql.DB) error {
	dbj.access.Lock()
	defer dbj.access.Unlock()
	dbj.tableName = "Formatteds"
	dbj.db = db
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS  Formatteds (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		fgroup TEXT,
		port INTEGER,
		boundType TEXT,
		protocolType TEXT,
		protocolSettings TEXT,
		mux INTEGER,
		streamSettings TEXT,
		proxyID INTEGER)
	`)
	if err != nil {
		return err
	}

	return nil
}
func (dbj *DataBaseFormatteds) Tags(m map[string]interface{}) (map[string][]string, error) {
	rows, err := dbj.Get(m)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		configMap, err := dbj.ToMap(rows)
		if err != nil {
			return nil, err
		}
		if configMap["boundType"] == "inbound" {
			return map[string][]string{
				"inbounds": []string{
					fmt.Sprintf(`{'type':'formatted','id':'%s'}`, configMap["id"].(string)),
				},
			}, nil
		} else if configMap["boundType"] == "outbound" {
			return map[string][]string{
				"outbounds": []string{
					fmt.Sprintf(`{'type':'formatted','id':'%d'}`, configMap["id"].(string)),
				},
			}, nil
		}
	}
	return nil, nil
}
func (dbj *DataBaseFormatteds) RegisterMap(m *map[string]DataBaseObj, db *sql.DB) {
	dbj.Init(db)
	(*m)["formatted"] = dbj
}
