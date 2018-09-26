package openape

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/buger/jsonparser"
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
		vHandler = func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			fmt.Printf("%s: %s \n", string(key), string(value))

			return nil
		}
		jsonparser.ObjectEach(body, vHandler)
		// TODO build insert request
		SendResponse(w, 200, map[string]string{"res": string(body)}, "application/json")
	}

}

// PostModels handles POST requests and inserts models
func (oape *OpenApe) PostModels(model string) []byte {

	return nil
}
