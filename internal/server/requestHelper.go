package server

import (
	"net/http"
	"strings"
)

var correctURLPathCount int = 5

func RequestCheck(res http.ResponseWriter, req *http.Request, contentType string) bool {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return false
	}

	/*	if req.Header.Get(`Content-Type`) != contentType {
		http.Error(res, "bad content-type", http.StatusBadRequest)
		return false
	}*/

	s := strings.Split(req.RequestURI, `/`)
	if len(s) != correctURLPathCount {
		http.Error(res, "not found", http.StatusNotFound)
		return false
	}

	return true
}

func ParseURL(url string) (string, string) {
	s := strings.Split(url, `/`)
	return s[3], s[4]
}
