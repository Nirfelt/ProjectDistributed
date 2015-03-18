package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

//string that points to the devise own home folder
//var basePath string = os.Getenv("HOME")

func main() {
	r := mux.NewRouter()

	update := r.Path("/hello").Subrouter()
	update.Methods("POST").HandlerFunc(HelloHandler)

	http.ListenAndServe(":9090", r)

}

func HelloHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Println("I hear you")

}
