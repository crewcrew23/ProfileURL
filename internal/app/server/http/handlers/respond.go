package handler

import (
	"encoding/json"
	"net/http"
)

func sendError(w http.ResponseWriter, code int, err error) {
	respond(w, code, map[string]string{"error": err.Error()})
}

func respond(w http.ResponseWriter, code int, data interface{}) {
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
