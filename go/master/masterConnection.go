package main

import (
	"fmt"
	//"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type dataNodes struct{
	node []string
}

func main() {
	nodes := dataNodes{}
	AddDataNode(nodes, "localhost:8080")
	AddDataNode(nodes, "localhost:8080")
	AddDataNode(nodes, "localhost:8080")
	r := mux.NewRouter()
	//s1 := r.Host(nodes.node[0]).Subrouter()
	//s2 := r.Host(nodes.node[0]).Subrouter()
	//s3 := r.Host(nodes.node[0]).Subrouter()
	file := r.Path("/{faculty}/{course}/{year}/{id}").Subrouter()
	file.Methods("GET").HandlerFunc(FileGetHandler)
	file.Methods("POST").HandlerFunc(FileCreateHandler)
	file.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	http.ListenAndServe(":8080", r)
}

func AddDataNode(nodes dataNodes, node string){
	nodes.node = append(nodes.node, node)
}

func FileCreateHandler(rw http.ResponseWriter, r *http.Request) {
	faculty := mux.Vars(r)["faculty"]
	course := mux.Vars(r)["course"]
	year := mux.Vars(r)["year"]
	id := mux.Vars(r)["id"]

	fmt.Fprintf(rw, "Created file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)
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

	fmt.Fprintf(rw, "Deleted file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)
}