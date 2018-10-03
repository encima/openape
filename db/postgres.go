package db

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/encima/openape/utils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // used for db connection
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// Database holds the db connection and sould implement support for creation, retrieval, adding, updating and deleting
type Database struct {
	Conn *sqlx.DB
}

const (
	baseCreationString string = "CREATE TABLE IF NOT EXISTS base_type (id VARCHAR PRIMARY KEY, created_at date, updated_at date);"
)

var (
	pgBaseTypes     = []string{"id", "created_at", "updated_at"}
	pgReservedWords = []string{"user", "group"}
)

// DatabaseConnect loads connection strings from the config file and connects to the specified DB
func DatabaseConnect() *sqlx.DB {
	engine, err := sqlx.Connect(viper.GetString("database.type"), viper.GetString("database.conn"))
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	return engine
}

// CreateSchema generates a creation string from a model and executes
func (db Database) CreateSchema(k string, props map[string]*openapi3.SchemaRef) {
	// Create parent table
	res, err := db.Conn.Exec(baseCreationString)
	if err != nil {
		fmt.Println(err)
		panic(fmt.Errorf("Problem creating BASE table %s", err))
	}
	fmt.Println(res)
	var createBytes strings.Builder
	createBytes.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", k))
	if utils.StringExists(k, pgReservedWords) {
		panic(fmt.Errorf("Reserved word found, table cannot be created"))
	}

	for k, v := range props {
		vType := v.Value.Type
		// remove fields that already exist in the `base` parent table
		if utils.StringExists(k, pgBaseTypes) {
			continue
		}
		dbType := "varchar"
		switch vType {
		case "integer":
			dbType = "integer"
			break
		case "object":
			dbType = "json"
			break
		case "boolean":
			dbType = "Boolean"
			break
		default:
			if v.Value.Format == "date-time" {
				dbType = "date"
			}
			break
		}
		createBytes.WriteString(fmt.Sprintf("%s %s", k, dbType))
		if k == "id" {
			createBytes.WriteString(" PRIMARY KEY,")
		} else {
			createBytes.WriteString(",")
		}
	}
	createStmt := createBytes.String()
	createBytes.Reset()
	createStmt = createStmt[:len(createStmt)-1]
	createStmt += ") INHERITS (base_type);"
	_, err = db.Conn.Exec(createStmt)
	if err != nil {
		panic(fmt.Errorf("Problem creating table for %s: %s", k, err))
	}
	fmt.Printf("Table %s created \n", k)
}

// GetModels queries a table of a model and returns all those that match
func (db Database) GetModels(w http.ResponseWriter, model string) {
	qString := fmt.Sprintf("SELECT * FROM %s", model)
	rows, err := db.Conn.Query(qString)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var v struct {
		Data []interface{} // `json:"data"`
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatal(err)
		}
		var m map[string]interface{}
		m = make(map[string]interface{})
		for i := range columns {
			m[columns[i]] = values[i]
		}
		v.Data = append(v.Data, m)
	}
	// jsonMsg, _ := json.Marshal(v)
	// TODO set content type from swagger and handle in method
	utils.SendResponse(w, 200, v, "application/json")
}

// PostModel finds the model to be created and inserts the record
func (db Database) PostModel(w http.ResponseWriter, modelName string, model *openapi3.SchemaRef, r *http.Request) {
	// r.ParseForm()
	// TODO only parse form when you know it is form, body is unreadable after this

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	if model != nil {
		reqKeys := model.Value.Required // all required properties of the matching model
		keyCount := 0
		for i := range reqKeys {
			_, dt, _, err := jsonparser.Get(body, reqKeys[i])
			if dt == jsonparser.NotExist || err != nil {
				msg := fmt.Sprintf("Required key '%s' is not present", reqKeys[i])
				e := map[string]string{"error": msg}
				utils.SendResponse(w, 400, e, "application/json")
				return
			}
		}

		var vHandler func([]byte, []byte, jsonparser.ValueType, int) error
		cols := make([]string, keyCount)
		vals := make([]interface{}, keyCount)
		vHandler = func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			fmt.Printf("%s: %s \n", string(key), string(value))
			keyCount++
			cols = append(cols, string(key))
			vals = append(vals, string(value))
			return nil
		}
		jsonparser.ObjectEach(body, vHandler)
		if !utils.StringExists("id", cols) {
			cols = append(cols, "id")
			u2 := uuid.Must(uuid.NewV4())
			vals = append(vals, u2.String())
		}
		cols = append(cols, "created_at")
		vals = append(vals, time.Now().Format(time.UnixDate))
		if len(cols) == len(vals) {
			var insertBytes strings.Builder
			var colsBytes strings.Builder
			var valsBytes strings.Builder
			insertBytes.WriteString(fmt.Sprintf("INSERT INTO %s (", modelName))
			index := 0
			for k := range cols {
				index++
				// TODO handle different data types here (encode json, quote strings, format datetimes etc)
				if index != len(cols) {
					colsBytes.WriteString(fmt.Sprintf("%s, ", cols[k]))
					valsBytes.WriteString(fmt.Sprintf("'%s', ", vals[k]))
				} else {
					colsBytes.WriteString(fmt.Sprintf("%s)", cols[k]))
					valsBytes.WriteString(fmt.Sprintf("'%s');", vals[k]))
				}
			}

			insertBytes.WriteString(fmt.Sprintf("%s VALUES (%s", colsBytes.String(), valsBytes.String()))
			fmt.Println(insertBytes.String())
			_, err := db.Conn.Exec(insertBytes.String())
			if err != nil {
				err := fmt.Sprintf("Problem inserting into table for %s: %s", modelName, err)
				utils.SendResponse(w, 404, map[string]string{"error": err}, "application/json")
				return
			}
			// TODO get ID and return here (or whole object?)
			utils.SendResponse(w, 200, map[string]string{"res": "Inserted successfully"}, "application/json")
			return
		}
		utils.SendResponse(w, 404, map[string]string{"error": "Object keys not equal to number of values"}, "application/json")
		return

	}

}
