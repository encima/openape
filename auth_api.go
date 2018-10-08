package openape

import (
	"fmt"
	"net/http"
)

// APIAuthHandler matches API key with user details stored
func (oape *OpenApe) APIAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO retrieve header key from swagger config
		token := r.Header.Get("X-API-KEY")
		sPath := r.Header.Get("X-OA-PATH")
		// apiPath := strings.Replace(r.URL.String(), "/api/v1", "", 1)
		print(sPath)
		swaggerPath := oape.swagger.Paths.Find(sPath)
		reqPath := swaggerPath.GetOperation(r.Method)
		print(reqPath)
		if reqPath.Security == nil {
			next.ServeHTTP(w, r)
		} else {
			for method := range *reqPath.Security {
				// TODO check security methods against those specified in swagger
				switch method {
				default:
					var apiKey string
					query := fmt.Sprintf("SELECT api_key FROM users where api_key='%s';", token)
					err := oape.db.Conn.Get(&apiKey, query)
					if err != nil {
						http.Error(w, "Forbidden", http.StatusForbidden)
					} else {
						next.ServeHTTP(w, r)
					}
					return
				}
			}

		}

	})
}
