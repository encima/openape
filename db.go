package openape

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // used for db connection
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

// GetModels queries a table of a model and returns all those that match
func (oape *OpenApe) GetModels(model string) []byte {
	qString := fmt.Sprintf("SELECT * FROM %s", model)
	rows, err := oape.db.Query(qString)
	if err != nil {
		fmt.Println(err)
	} else {
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
		jsonMsg, _ := json.Marshal(v)
		return jsonMsg
	}
	return nil
}

// PostModel finds the model to be created and inserts the record
func (oape *OpenApe) PostModel(model string, r *http.Request) []byte {
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Printf("%s: %s", k, v)
	}
	// TODO get model and values associated with it, extract from request using FormValue(<field>)
	// TODO build insert request
	return nil
}
