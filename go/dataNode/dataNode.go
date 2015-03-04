package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

//string that points to the devise own home folder
var basePath string = os.Getenv("HOME")

func main() {
	r := mux.NewRouter()
	file := r.Path("/{faculty}/{course}/{year}/{id}").Subrouter()
	file.Methods("GET").HandlerFunc(FileGetHandler)
	file.Methods("POST").HandlerFunc(FileCreateHandler)
	file.Methods("DELETE").HandlerFunc(FileDeleteHandler)
	uFile := r.Path("/upload/{id}").Subrouter()
	uFile.Methods("POST").HandlerFunc(UploadHandler)

	http.ListenAndServe(":8080", r)
}

func FileCreateHandler(rw http.ResponseWriter, r *http.Request) {
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
	fmt.Fprintf(rw, "File path: %s", filePath)
}

func FileGetHandler(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

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

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	filePath := path.Join(basePath, id)
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	defer file.Close()

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing.")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

//Function to get all files from another data node

//Function when a data node has been down and starts over it contacts web server to get ip to master to contact.

//Function to contact master to get an ip to another data node
