package main

import (
	"fmt"
	//"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
	//"os"
	"net/url"
)

type dataNodes struct{
	node []node
}

type node struct{
	address string
	ok bool
}

var nodes = dataNodes{} // List with dataNodes (struct)

func main() {
	//Declare functions
	AddDataNode("localhost:8080")
	AddDataNode("localhost:8080")
	AddDataNode("localhost:8080")
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

func AddDataNode(address string){
	node := node{address: address, ok: false}
	nodes.node = append(nodes.node, node)
	//update DB
}

func RemoveDataNode(node string){
	if len(nodes.node) == 0 {
		return
	}
	for i := range nodes.node{
		if nodes.node[i].address == node {
			nodes.node[i] = nodes.node[len(nodes.node)-1]
			nodes.node = nodes.node[:len(nodes.node)-1]
		}
	}
	//Update DB
}

func FileCreateHandler(rw http.ResponseWriter, r *http.Request) {
	//Multicast post file to all datanodes
	for i := range nodes.node{
		if nodes.node[i].ok == true{
			resp, err := http.PostForm(nodes.node[i].address, url.Values{"key": {"Value"}, "id": {"123"}})
			if err != nil {
				fmt.Println("ERROR")
			}
			defer resp.Body.Close()
		}
	}
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

//func get datanode ip 

//func return all files and folders