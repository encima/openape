package openape

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Jeffail/gabs"
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
	// r.ParseForm()
	// TODO only parse form when you know it is form, body is unreadable after this
	m := oape.swagger.Components.Schemas[model]
	body, err := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	for i := range m.Value.Properties {
		print(i)
		// TODO look for keys here and see if they match the object
		// TODO split into validation function (utils)
	}

	value, _ := jsonParsed.Path("name").Data().(string)
	print(value)
	/*if m != nil {
		reqKeys := m.Value.Required // all required properties of the matching model
		for k, v := range r.Form {
			fmt.Printf("%s: %s", k, v)
			for mk, _ := range m.Value.Properties {
				if k == mk && StringExists(k, reqKeys) {
					fmt.Println("Found required")
				}
			}
		}
	}*/

	// TODO get model and values associated with it, extract from request using FormValue(<field>)
	// TODO build insert request
	return nil
}
