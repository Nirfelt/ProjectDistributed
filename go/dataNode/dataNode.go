package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
)

var basePath string = os.Getenv("HOME")

func main() {
	r := mux.NewRouter()
	file := r.Path("/{faculty}/{course}/{year}/{id}").Subrouter()
	file.Methods("GET").HandlerFunc(FileGetHandler)
	file.Methods("POST").HandlerFunc(FileCreateHandler)
	file.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	http.ListenAndServe(":8080", r)
}

func FileCreateHandler(rw http.ResponseWriter, r *http.Request) {
	faculty := mux.Vars(r)["faculty"]
	course := mux.Vars(r)["course"]
	year := mux.Vars(r)["year"]
	id := mux.Vars(r)["id"]

	filePath := path.Join(basePath, id)
	body, _ := ioutil.ReadAll(r.Body)

	if _, err := os.Stat(filePath); err == nil {
		rw.WriteHeader(http.StatusConflict)
		rw.Write([]byte("File exists"))
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error: " + err.Error()))
		return
	}
	defer file.Close()

	file.Write(body)
	file.Sync()

	rw.WriteHeader(http.StatusOK)
	log.Printf("Created file with id: %s\n", id)

	fmt.Fprintf(rw, "Created file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)
	fmt.Fprintf(rw, "File path: %s", filePath)
}

func FileGetHandler(rw http.ResponseWriter, r *http.Request) {
	faculty := mux.Vars(r)["faculty"]
	course := mux.Vars(r)["course"]
	year := mux.Vars(r)["year"]
	id := mux.Vars(r)["id"]

	fmt.Fprintf(rw, "Get file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)

}

func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	faculty := mux.Vars(r)["faculty"]
	course := mux.Vars(r)["course"]
	year := mux.Vars(r)["year"]
	id := mux.Vars(r)["id"]

	fmt.Fprintf(rw, "Delete file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)

}
