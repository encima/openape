package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type OpenApe struct {
	db      *xorm.Engine
	router  *mux.Router
	swagger *openapi3.Swagger
}

// RootHandler responds to / request
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("go APE!\n"))
}

// LoadConfig loads config file using Viper package
func LoadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

// AddRoute takes a path and a method to create a route handler for a Mux router instance
func AddRoute(r *mux.Router, path string, method string) {
	r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(path))
	}).Methods(method)
}

// MapModels reads the models from the provided swagger file and creates the correspdonding tables in Postgres
func MapModels(models map[string]*openapi3.SchemaRef, o OpenApe) {
	// Create parent table
	_, err := o.db.Exec(viper.GetString("system-db.parent-stmt"))
	if err != nil {
		panic(fmt.Errorf("Problem creating BASE table %s", err))
	}
	for k, v := range models {

		tableInsert := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", k)

		w := viper.Get("system-db.reserved-words").([]interface{})
		words := ToStringMap(w)
		if StringExists(k, words) {
			panic(fmt.Errorf("Reserved word found, table cannot be created"))
		}

		for k, v := range v.Value.Properties {
			vType := v.Value.Type
			// remove fields that already exist in the `base` parent table
			words = ToStringMap(viper.Get("system-db.parent-fields").([]interface{}))
			if StringExists(k, words) {
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
			tableInsert += fmt.Sprintf("%s %s", k, dbType)
			if k == "id" {
				tableInsert += " PRIMARY KEY,"
			} else {
				tableInsert += ","
			}
		}
		tableInsert = tableInsert[:len(tableInsert)-1]
		tableInsert += ") INHERITS (base);"
		_, err := o.db.Exec(tableInsert)
		if err != nil {
			panic(fmt.Errorf("Problem creating table for %s: %s", k, err))
		}
		fmt.Printf("Table %s created \n", k)
	}
}

// MapRoutes iterates the paths laid out in the swagger file and adds them to the router
func MapRoutes(paths map[string]*openapi3.PathItem, o OpenApe) {
	for k, v := range paths {
		fmt.Println(k)
		if op := v.GetOperation("GET"); op != nil {
			AddRoute(o.router, k, "GET")
		}
		if op := v.GetOperation("PUT"); op != nil {

		}
		if op := v.GetOperation("POST"); op != nil {

		}
		if op := v.GetOperation("DELETE"); op != nil {

		}
	}
}

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	LoadConfig()
	r.HandleFunc("/", RootHandler).Methods("GET")

	staticDir := viper.GetString("server.static")

	dbEngine := DatabaseConnect()

	oapiPath := viper.GetString("openapi.path")
	swagger := LoadSwagger(oapiPath)

	o := OpenApe{dbEngine, r, swagger}

	o.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	MapModels(swagger.Components.Schemas, o)
	MapRoutes(swagger.Paths, o)

	// Bind to a port and pass our router in
	port := fmt.Sprintf(":%s", viper.GetString("server.port"))
	log.Fatal(http.ListenAndServe(port, r))
}
