package openape

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // used for db connection
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// DatabaseConnect loads connection strings from the config file and connects to the specified DB
func DatabaseConnect() *sqlx.DB {
	engine, err := sqlx.Connect(viper.GetString("database.type"), viper.GetString("database.conn"))
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	return engine
}

// CreateTable generates a creation string from a model and executes
func (oape *OpenApe) CreateTable(k string, props map[string]*openapi3.SchemaRef) {
	var createBytes strings.Builder
	createBytes.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", k))
	if StringExists(k, pgReservedWords) {
		panic(fmt.Errorf("Reserved word found, table cannot be created"))
	}

	for k, v := range props {
		vType := v.Value.Type
		// remove fields that already exist in the `base` parent table
		if StringExists(k, pgBaseTypes) {
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
	_, err := oape.db.Exec(createStmt)
	if err != nil {
		panic(fmt.Errorf("Problem creating table for %s: %s", k, err))
	}
	fmt.Printf("Table %s created \n", k)
}

// GetModels queries a table of a model and returns all those that match
func (oape *OpenApe) GetModels(w http.ResponseWriter, model string) {
	qString := fmt.Sprintf("SELECT * FROM %s", model)
	rows, err := oape.db.Query(qString)
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
	SendResponse(w, 200, v, "application/json")
}

// PostModel finds the model to be created and inserts the record
func (oape *OpenApe) PostModel(w http.ResponseWriter, model string, r *http.Request) {
	// r.ParseForm()
	// TODO only parse form when you know it is form, body is unreadable after this
	m := oape.swagger.Components.Schemas[model]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	if m != nil {
		reqKeys := m.Value.Required // all required properties of the matching model
		keyCount := 0
		for i := range reqKeys {
			_, dt, _, err := jsonparser.Get(body, reqKeys[i])
			if dt == jsonparser.NotExist || err != nil {
				msg := fmt.Sprintf("Required key '%s' is not present", reqKeys[i])
				e := map[string]string{"error": msg}
				SendResponse(w, 400, e, "application/json")
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
		if !StringExists("id", cols) {
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
			insertBytes.WriteString(fmt.Sprintf("INSERT INTO %s (", model))
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
			_, err := oape.db.Exec(insertBytes.String())
			if err != nil {
				err := fmt.Sprintf("Problem inserting into table for %s: %s", model, err)
				SendResponse(w, 404, map[string]string{"error": err}, "application/json")
				return
			}
			// TODO get ID and return here (or whole object?)
			SendResponse(w, 200, map[string]string{"res": "Inserted successfully"}, "application/json")
			return
		}
		SendResponse(w, 404, map[string]string{"error": "Object keys not equal to number of values"}, "application/json")
		return

	}

}

// PostModels handles POST requests and inserts models
func (oape *OpenApe) PostModels(model string) []byte {

	return nil
}
