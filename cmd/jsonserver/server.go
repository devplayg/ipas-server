package main

import (
	"net/http"
	"encoding/json"
	"fmt"
)

func main() {
	http.HandleFunc("/event", handle)
	http.ListenAndServe(":8000", nil)
}

func handle(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	jsonMap := make(map[string]interface{})
	err := decoder.Decode(&jsonMap)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	jsonRes, err := json.MarshalIndent(jsonMap, "", "    ")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		fmt.Fprintf(w, string(jsonRes))
	}

}
