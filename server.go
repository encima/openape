package openape

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

// OpenApe object to hold objects related to the server
type OpenApe struct {
	db      *sqlx.DB
	router  *mux.Router
	swagger *openapi3.Swagger
	config  *viper.Viper
}

// JSONResponse is alias of map for JSON response
type JSONResponse struct {
	data   map[string]interface{}
	status int
}

const (
	baseCreationString string = "CREATE TABLE IF NOT EXISTS base_type (id VARCHAR PRIMARY KEY, created_at date, updated_at date);"
)

var (
	pgBaseTypes     = []string{"id", "created_at", "updated_at"}
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
func (oape *OpenApe) AddRoute(path string, method string, model string) {
	fmt.Printf("Adding route: %s \n", path)
	oape.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var res []byte
		switch method {
		case "GET":
			res = oape.GetModels(model)
			break
		case "POST":
			res = oape.PostModel(model, r)
			break
		default:
			break
		}
		w.Write(res)
	}).Methods(method)
}

// MapModels reads the models from the provided swagger file and creates the correspdonding tables in Postgres
func (oape *OpenApe) MapModels(models map[string]*openapi3.SchemaRef) {
	// Create parent table
	res, err := oape.db.Exec(baseCreationString)
	if err != nil {
		fmt.Println(err)
		panic(fmt.Errorf("Problem creating BASE table %s", err))
	}
	fmt.Println(res)

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
		tableInsert += ") INHERITS (base_type);"
		_, err := oape.db.Exec(tableInsert)
		if err != nil {
			panic(fmt.Errorf("Problem creating table for %s: %s", k, err))
		}
		fmt.Printf("Table %s created \n", k)
	}
}

// GetModelFromPath identifies which routes maps to which models identified in the Schemas of the spec
func (oape *OpenApe) GetModelFromPath(path string) string {
	for k := range oape.swagger.Components.Schemas {
		if strings.Contains(strings.ToLower(path), strings.ToLower(k)) {
			return k
		}
	}
	return ""
}

// MapRoutes iterates the paths laid out in the swagger file and adds them to the router
func (oape *OpenApe) MapRoutes(paths map[string]*openapi3.PathItem) {
	for k, v := range paths {
		model := oape.GetModelFromPath(k)
		if op := v.GetOperation("GET"); op != nil {
			oape.AddRoute(k, "GET", model)
		}
		if op := v.GetOperation("PUT"); op != nil {
			oape.AddRoute(k, "PUT", model)
		}
		if op := v.GetOperation("POST"); op != nil {
			oape.AddRoute(k, "POST", model)
		}
		if op := v.GetOperation("DELETE"); op != nil {
			oape.AddRoute(k, "DELETE", model)
		}
	}
}

// RunServer starts the openapi server on the specified port
func (oape *OpenApe) RunServer() {
	port := ":8080" // Following the open api docs, the default URL should be /
	// if len(oape.swagger.Servers) > 0 {
	// 	port = oape.swagger.Servers[0].Variables["port"].Default.(string)
	// }
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
	if len(o.swagger.Servers) > 0 && o.swagger.Servers[0].Variables["basePath"] != nil {
		o.router = o.router.PathPrefix("/api/v1").Subrouter()
	}

	// set up with routes and models to DB and Router
	o.MapModels(swagger.Components.Schemas)
	o.MapRoutes(swagger.Paths)

	return o
}
