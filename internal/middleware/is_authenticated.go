package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func IsAuthenticated(ctx context.Context) func(next http.Handler) http.Handler {
	return Handler(ctx)
}

func Handler(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			errorMessage := "Erro na autenticação"
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				err := errors.New("Unauthorized")
				respondWithError(rw, http.StatusUnauthorized, err.Error(), errorMessage)
				return
			}
			payload := `{
				"token": "` + tokenString + `"
			}`
			req, err := http.Post(os.Getenv("AUTH_URL")+"/v1/validate-token", "text/plain", strings.NewReader(payload))
			if err != nil {
				respondWithError(rw, http.StatusUnauthorized, err.Error(), errorMessage)
				return
			}
			defer req.Body.Close()

			type result struct {
				Email string `json:"email"`
			}
			var res result
			err = json.NewDecoder(req.Body).Decode(&res)
			if err != nil {
				respondWithError(rw, http.StatusUnauthorized, err.Error(), errorMessage)
				return
			}
			newCTX := context.WithValue(r.Context(), "email", res.Email)
			next.ServeHTTP(rw, r.WithContext(newCTX))
		}
		return http.HandlerFunc(fn)
	}
}

// RespondWithError return a http error
func respondWithError(w http.ResponseWriter, code int, e string, message string) {
	respondWithJSON(w, code, map[string]string{"code": strconv.Itoa(code), "error": e, "message": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
