package openape

import (
	"fmt"
	"net/http"
)

// TokenAuthMiddleware takes the `X-API-KEY` header and searches for a match in the database
func (oape *OpenApe) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-KEY")
		var users interface{}
		count, err := oape.db.Table("users").Where("api_key == %s", token).Count(&users)
		if err != nil {
			panic(fmt.Errorf("Error authenticating user: %s", err))
		}
		if count == 1 {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
