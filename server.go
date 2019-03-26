package openape

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Jumpscale/go-raml/raml"

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
	ramlAPI *raml.APIDefinition
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

// AddCustomRoute is for those routes that need handling beyond the basic CRUD operations
func (oape *OpenApe) AddCustomRoute(path string, method string, handler func(w http.ResponseWriter, r *http.Request)) {
	oape.router.HandleFunc(path, handler).Methods(method)
}

// AddCRUDRoute takes a path and a method to create a route handler for a Mux router instance
func (oape *OpenApe) AddCRUDRoute(path string, method string, model string) {
	fmt.Printf("Adding route: %s %s \n", method, path)
	oape.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var res utils.JSONResponse
		switch method {
		case "GET":
			res = oape.db.GetModels(model)
			break
		case "POST":
			m := oape.swagger.Components.Schemas[model]
			// TODO default behaviour is to pass the model to the db and create. Needs to handle special methods (Login etc)
			res = oape.db.PostModel(model, m, r)
			break
		case "PUT":
			m := oape.swagger.Components.Schemas[model]
			res = oape.db.PutModel(vars["id"], model, m, r)
			break
		case "DELETE":
			res = oape.db.DeleteModel(vars["id"], model, r)
			break
		default:
			break
		}
		utils.SendResponse(w, res)
	}).Methods(method)
}

// RunServer starts the openapi server on the specified port
func (oape *OpenApe) RunServer() {
	port := ":8080"
	if oape.swagger != nil && len(oape.swagger.Servers) > 0 {
		port = oape.swagger.Servers[0].Variables["port"].Default.(string)
	} else if oape.ramlAPI != nil {
		port = fmt.Sprintf(":%d", oape.ramlAPI.BaseURIParameters["port"].Default)
	}

	fmt.Printf("Server running on port %s \n", port)
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

	ramlAPI := new(raml.APIDefinition)
	var swagger *openapi3.Swagger
	ramlPath := viper.GetString("raml.path")
	if len(ramlPath) > 0 {
		err := raml.ParseFile(ramlPath, ramlAPI)
		fmt.Println("Loading RAML spec")
		if err != nil {
			panic(fmt.Errorf("%s", err))
		}
	} else {
		ramlAPI = nil
		oapiPath := viper.GetString("openapi.path")
		oapiSrc := viper.GetString("openapi.src")
		swagger = utils.LoadSwagger(oapiPath, oapiSrc)
		fmt.Println("Loading OpenAPI spec")
	}
	odb := db.Database{Conn: dbEngine}
	o := OpenApe{odb, r, swagger, viper.GetViper(), ramlAPI}
	o.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	// TODO handle base path from config files
	o.router = o.router.PathPrefix("/api/v1").Subrouter()
	o.router.Use(o.APIAuthHandler)
	if o.ramlAPI != nil {
		fmt.Println("Loading RAML specification...")
		o.MapRAMLModels()
		res := make(map[string]*raml.Resource)
		for k, v := range o.ramlAPI.Resources {
			res[k] = &v
		}
		o.MapRAMLResources(res)
	} else if o.swagger != nil {
		fmt.Println("Loading OpenAPI (3) specification...")
		o.MapModels(swagger.Components.Schemas)
		o.MapRoutes(swagger.Paths)
	} else {
		panic("No API has been provided")
	}

	return o
}
