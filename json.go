package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJson(w http.ResponseWriter, r *http.Request, a interface{}) (err error) {
	callbackFuncName := r.FormValue("callback")
	data, err := json.Marshal(a)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if callbackFuncName != "" {
			//	関数の存在確認付き
			fmt.Fprintf(w, "%s && %s(%s);", callbackFuncName, callbackFuncName, string(data))
		} else {
			fmt.Fprintf(w, "%s", string(data))
		}
	}
	return
}
