package main

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/lib/pq"
)

func checkIfNil(err interface{}) bool {
	var result bool = err != nil
	if result {
		panic(err)
	}

	return result
}

func initConfigs() {
	rf, err := os.ReadFile("./protected/config.json")
	checkIfNil(err)

	var configJSON configStruct

	errjson := configJSON.UnmarshalJSON(rf)
	checkIfNil(errjson)
	globalruntimeparams.singlemode = configJSON.singlemode
	globalruntimeparams.aviliablefiles = configJSON.aviliablefiles

	connStr := "user=" + configJSON.database.user + " password=" + configJSON.database.password + " dbname=" + configJSON.database.dbname + " sslmode=" + configJSON.database.sslmode + ""
	dbpg, errpg := sql.Open(configJSON.database.driver, connStr)
	checkIfNil(errpg)
	globalruntimeparams.driver = dbpg
}

type configStruct struct {
	singlemode     bool
	database       databaseStruct
	aviliablefiles []string
}

type databaseStruct struct {
	driver   string
	user     string
	password string
	dbname   string
	sslmode  string
}

func (cs *configStruct) UnmarshalJSON(b []byte) error {
	var tmp struct {
		Singlemode     bool           `json:"singlemode"`
		Database       databaseStruct `json:"database"`
		Aviliablefiles []string       `json:"aviliablefiles"`
	}

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	cs.database = tmp.Database

	return nil
}

func (cs *databaseStruct) UnmarshalJSON(b []byte) error {
	var tmp struct {
		Driver   string `json:"driver"`
		User     string `json:"user"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
		Sslmode  string `json:"sslmode"`
	}

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	cs.driver = tmp.Driver
	cs.user = tmp.User
	cs.password = tmp.Password
	cs.dbname = tmp.Dbname
	cs.sslmode = tmp.Sslmode

	return nil
}

func closeAllConnections() {

	globalruntimeparams.driver.Close()
	for _, token := range voiceSessionMaster {

		if token.session != nil {
			token.session.Close()
		}
	}
}

type runtimeparams struct {
	singlemode     bool
	driver         *sql.DB
	aviliablefiles []string
}
