package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
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
func LoadSwagger(p string, src string) *openapi3.Swagger {

	var swagger *openapi3.Swagger
	var err error
	switch src {
	case "FILE":
		swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromFile(p)
		break
	case "URL":
		u, err := url.Parse(p)
		if err != nil {
			panic(fmt.Errorf("Error parsing URL: %s", err))
		}
		swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromURI(u)
		break
	default:
		break
	}

	if err != nil {
		panic(fmt.Errorf("Error reading swagger file: %s", err))
	}

	return swagger
}

// SendResponse handles writing json responses back to the client using http response writer
func SendResponse(w http.ResponseWriter, res JSONResponse) {
	w.Header().Set("Content-Type", res.ContentType)
	w.WriteHeader(res.Status)
	w.Write(res.Data)

}

// GetBytesFromInterface returns a byte array for the specified generic interface
func GetBytesFromInterface(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
