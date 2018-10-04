package openape

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/encima/openape/db"
	"github.com/encima/openape/utils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// OpenApe object to hold objects related to the server
type OpenApe struct {
	db      db.Database
	router  *mux.Router
	swagger *openapi3.Swagger
	config  *viper.Viper
}

const (
	baseCreationString string = "CREATE TABLE IF NOT EXISTS base_type (id VARCHAR PRIMARY KEY, created_at date, updated_at date);"
)

var (
	pgBaseTypes     = []string{"id", "created_at", "updated_at"}
	pgReservedWords = []string{"user", "group"}
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
		switch method {
		case "GET":
			oape.db.GetModels(w, model)
			break
		case "POST":
			m := oape.swagger.Components.Schemas[model]
			oape.db.PostModel(w, model, m, r)
			break
		case "PUT":
			m := oape.swagger.Components.Schemas[model]
			vars := mux.Vars(r)
			oape.db.PutModel(w, vars["id"], model, m, r)
			break
		default:
			break
		}
	}).Methods(method)
}

// MapModels reads the models from the provided swagger file and creates the correspdonding tables in Postgres
func (oape *OpenApe) MapModels(models map[string]*openapi3.SchemaRef) {
	for k, v := range models {
		oape.db.CreateSchema(k, v.Value.Properties)
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
		// TODO handle when user specifies function and do not pass to route
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
	port := ":8080"
	if len(oape.swagger.Servers) > 0 {
		port = oape.swagger.Servers[0].Variables["port"].Default.(string)
	}
	log.Fatal(http.ListenAndServe(port, oape.router))
}

// NewServer sets up the
func NewServer(configPath string) OpenApe {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	LoadConfig(configPath)
	r.HandleFunc("/", RootHandler).Methods("GET")

	staticDir := viper.GetString("server.static")

	dbEngine := db.DatabaseConnect()

	oapiPath := viper.GetString("openapi.path")
	swagger := utils.LoadSwagger(oapiPath)

	odb := db.Database{dbEngine}
	o := OpenApe{odb, r, swagger, viper.GetViper()}
	o.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	if len(o.swagger.Servers) > 0 && o.swagger.Servers[0].Variables["basePath"] != nil {
		o.router = o.router.PathPrefix("/api/v1").Subrouter()
	}

	// set up with routes and models to DB and Router
	o.MapModels(swagger.Components.Schemas)
	o.MapRoutes(swagger.Paths)

	return o
}
