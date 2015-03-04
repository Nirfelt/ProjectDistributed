package main

import (
	"fmt"
	//"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
)

type dataNodes struct{
	node []node
}

type node struct{
	address string
	ok bool
}

func main() {
	nodes := dataNodes{}
	nodes = AddDataNode(nodes, "localhost:8080")
	nodes = AddDataNode(nodes, "localhost:8080")
	nodes = AddDataNode(nodes, "localhost:8080")
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

func AddDataNode(nodes dataNodes, address string) dataNodes{
	node := node{address: address, ok: false}
	nodes.node = append(nodes.node, node)
	return nodes
}

func RemoveDataNode(nodes dataNodes, node string) dataNodes{
	if len(nodes.node) == 0 {
		return nodes
	}
	for i := range nodes.node{
		if nodes.node[i].address == node {
			nodes.node[i] = nodes.node[len(nodes.node)-1]
			nodes.node = nodes.node[:len(nodes.node)-1]
		}
	}
	return nodes
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