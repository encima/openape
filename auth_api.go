package openape

import (
	"fmt"
	"net/http"
	"strings"
)

// APIAuthHandler matches API key with user details stored
func (oape *OpenApe) APIAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO retrieve header key from swagger config
		token := r.Header.Get("X-API-KEY")
		apiPath := strings.Replace(r.RequestURI, "/api/v1", "", 1)

		if oape.Swagger != nil {
			swaggerPath := oape.Swagger.Paths.Find(apiPath)
			reqPath := swaggerPath.GetOperation(r.Method)
			if reqPath.Security == nil {
				next.ServeHTTP(w, r)
			} else if token != "" {
				for method := range *reqPath.Security {
					// TODO check security methods against those specified in swagger
					switch method {
					default:
						var apiKey string
						query := fmt.Sprintf("SELECT api_key FROM users where api_key='%s';", token)
						err := oape.DB.Conn.Get(&apiKey, query)
						if err != nil {
							http.Error(w, "Forbidden", http.StatusForbidden)
						} else {
							next.ServeHTTP(w, r)
						}
						return
					}
				}
			} else {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
			if oape.RamlAPI != nil {
				apiPath := strings.Replace(r.URL.String(), "/api/v1", "", 1)
				ramlPath := oape.RamlAPI.Resources[apiPath]
				method := ramlPath.MethodByName(r.Method)
				fmt.Println(method)
				next.ServeHTTP(w, r)
			}
		}
	})
}
