package main

import (
	"fmt"
	//"io"
	"net/http"
	"github.com/gorilla/mux"
	//"os"
	"net/url"
	//"bytes"
	"flag"
	"log"
	//"mime/multipart"
	"unicode"
	"strings"
)

type dataNodes struct{
	node []node
}

type node struct{
	address string
	ok bool
}

var nodes = dataNodes{} // List with dataNodes (struct)

var (
	listen = flag.String("listen", "localhost:8080", "listen on address")
	logp = flag.Bool("log", false, "enable logging")
)

func main() {
	//Declare functions
	flag.Parse()
	AddDataNode("localhost:8081")
	//AddDataNode("localhost:8082")
	//r := mux.NewRouter()
	//update := r.Path("/update")//.Subrouter()
	//update.Methods("POST").HandlerFunc(FileUploadHandler)
	proxyHandler := http.HandlerFunc(proxyHandlerFunc)
	log.Fatal(http.ListenAndServe(*listen, proxyHandler))

	//err := http.ListenAndServe(":8080", r)
	//if err != nil{
	//	log.Fatal("ListenAndServe: ", err)
	//}
}

func AddDataNode(address string){
	node := node{address: address, ok: true}
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

func proxyHandlerFunc(rw http.ResponseWriter, r *http.Request) {

	for i := 0; i < len(nodes.node); i++ {
		req := r
		client := &http.Client{}
		req.RequestURI = ""

		u, err := url.Parse("http://" + nodes.node[i].address + "/update")
	    if err != nil {
	        panic(err)
	    }   
	    req.URL = u
	    fmt.Println(u.String())
		req.URL.Scheme = strings.Map(unicode.ToLower, req.URL.Scheme)
		// And proxy
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Write(rw)
	}
	
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