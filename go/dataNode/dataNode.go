package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	//"mime/multipart"
	"net/http"
	"os"
	"path"
)

//string that points to the devise own home folder
var basePath string = os.Getenv("HOME")

func main() {
	r := mux.NewRouter()
	file := r.Path("/{id}").Subrouter()
	file.Methods("GET").HandlerFunc(FileGetHandler)
	file.Methods("POST").HandlerFunc(FileUploadHandler)
	file.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	http.ListenAndServe(":8080", r)
}

func FileGetHandler(rw http.ResponseWriter, r *http.Request) {
	//Takes the id from url
	id := mux.Vars(r)["id"]

	//set file path
	filePath := path.Join(basePath, id)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(data)

}

func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	filePath := path.Join(basePath, id)

	err := os.Remove(filePath)
	if err != nil {
		// Better error handling would be nice..
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
	log.Printf("Deleted file with id: %s\n", id)
}

func FileUploadHandler(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	faculty := r.FormValue("faculty")
	course := r.FormValue("course")
	year := r.FormValue("year")

	fmt.Fprintf(rw, "Faculty: %s, Course: %s, Year: %s", faculty, course, year)
	// the FormFile function takes in the POST input id file
	file, _, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	filePath := path.Join(basePath, id)

	defer file.Close()

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(rw, "Unable to create the file for writing.")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(rw, err)
	}

	fmt.Fprintf(rw, "File uploaded successfully: %s\n", id)
}

//Function to get all files from another data node

//Function when a data node has been down and starts over it contacts web server to get ip to master to contact.

//Function to contact master to get an ip to another data node
