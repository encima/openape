package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// JSONResponse is alias of map for JSON response
type JSONResponse struct {
	Data        []byte
	Status      int
	ContentType string
}

// StringExists checks for a needle (string) in a haystack (array)
func StringExists(needle string, haystack []string) bool {
	for _, h := range haystack {
		if strings.ToLower(needle) == strings.ToLower(h) {
			return true
		}
	}
	return false
}

// ToStringMap converts YAML arrays to their source string array
func ToStringMap(source []interface{}) []string {
	target := make([]string, len(source))
	for i := 0; i < len(source); i++ {
		target[i] = source[i].(string)
	}
	return target
}

// LoadSwagger loads swagger from specified location
func LoadSwagger(p string) *openapi3.Swagger {

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(p)
	if err != nil {
		panic(fmt.Errorf("Error reading swagger file: %s", err))
	}

	fmt.Println(swagger.Servers[0])
	return swagger
}

// SendResponse handles writing json responses back to the client using http response writer
func SendResponse(w http.ResponseWriter, res JSONResponse) {
	w.Header().Set("Content-Type", res.ContentType)
	w.WriteHeader(res.Status)
	w.Write(res.Data)

}
