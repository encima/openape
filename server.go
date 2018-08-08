package openape

import (
	"fmt"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// OpenApe object to hold objects related to the server
type OpenApe struct {
	db      *xorm.Engine
	router  *mux.Router
	swagger *openapi3.Swagger
	config  *viper.Viper
}

const (
	baseCreationString string = "CREATE TABLE IF NOT EXISTS base (id VARCHAR PRIMARY KEY, created_at date, modified_at date);"
)

var (
	pgBaseTypes     = []string{"id", "created_at", "modified_at"}
	pgReservedWords = []string{"user"}
)

// RootHandler responds to / request
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("go APE!\n"))
}

// LoadConfig loads config file using Viper package
func LoadConfig(path string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(path)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

// AddRoute takes a path and a method to create a route handler for a Mux router instance
func (oape *OpenApe) AddRoute(path string, method string) {
	oape.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(path))
	}).Methods(method)
}

// MapModels reads the models from the provided swagger file and creates the correspdonding tables in Postgres
func (oape *OpenApe) MapModels(models map[string]*openapi3.SchemaRef) {
	// Create parent table
	_, err := oape.db.Exec(baseCreationString)
	if err != nil {
		panic(fmt.Errorf("Problem creating BASE table %s", err))
	}
	for k, v := range models {

		tableInsert := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", k)
		if StringExists(k, pgReservedWords) {
			panic(fmt.Errorf("Reserved word found, table cannot be created"))
		}

		for k, v := range v.Value.Properties {
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
			tableInsert += fmt.Sprintf("%s %s", k, dbType)
			if k == "id" {
				tableInsert += " PRIMARY KEY,"
			} else {
				tableInsert += ","
			}
		}
		tableInsert = tableInsert[:len(tableInsert)-1]
		tableInsert += ") INHERITS (base);"
		_, err := oape.db.Exec(tableInsert)
		if err != nil {
			panic(fmt.Errorf("Problem creating table for %s: %s", k, err))
		}
		fmt.Printf("Table %s created \n", k)
	}
}

// MapRoutes iterates the paths laid out in the swagger file and adds them to the router
func (oape *OpenApe) MapRoutes(paths map[string]*openapi3.PathItem) {
	for k, v := range paths {
		fmt.Println(k)
		if op := v.GetOperation("GET"); op != nil {
			oape.AddRoute(k, "GET")
		}
		if op := v.GetOperation("PUT"); op != nil {
			oape.AddRoute(k, "GET")
		}
		if op := v.GetOperation("POST"); op != nil {
			oape.AddRoute(k, "GET")
		}
		if op := v.GetOperation("DELETE"); op != nil {
			oape.AddRoute(k, "GET")
		}
	}
}

// RunServer starts the openapi server on the specified port
func (oape *OpenApe) RunServer() {
	port := fmt.Sprintf(":%s", oape.config.GetString("server.port"))
	log.Fatal(http.ListenAndServe(port, oape.router))
}

// NewServer sets up the
func NewServer(configPath string) OpenApe {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	LoadConfig(configPath)
	r.HandleFunc("/", RootHandler).Methods("GET")

	staticDir := viper.GetString("server.static")

	dbEngine := DatabaseConnect()

	oapiPath := viper.GetString("openapi.path")
	swagger := LoadSwagger(oapiPath)

	o := OpenApe{dbEngine, r, swagger, viper.GetViper()}

	o.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// set up with routes and models to DB and Router
	o.MapModels(swagger.Components.Schemas)
	o.MapRoutes(swagger.Paths)

	return o
}
