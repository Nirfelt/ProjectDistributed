package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"unicode"
)

type dataNodes struct {
	node []node
}

type node struct {
	address string
	ok      bool
}

var nodes = dataNodes{} // List with dataNodes (struct)

var (
	listen = flag.String("listen", "localhost:8080", "listen on address")
	logp   = flag.Bool("log", false, "enable logging")
)

func main() {
	//Declare functions
	flag.Parse()
	AddDataNode("localhost:8081")
	AddDataNode("localhost:8082")
	AddDataNode("localhost:8083")
	r := mux.NewRouter()
	update := r.Path("/update")
	update.Methods("POST").HandlerFunc(ProxyHandlerFunc)
	handshake := r.Path("/handshake")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)
	http.ListenAndServe(":8080", r)
}

func AddDataNode(address string) {
	node := node{address: address, ok: true}
	nodes.node = append(nodes.node, node)
	//update DB
}

func RemoveDataNode(node string) {
	if len(nodes.node) == 0 {
		return
	}
	for i := range nodes.node {
		if nodes.node[i].address == node {
			nodes.node[i] = nodes.node[len(nodes.node)-1]
			nodes.node = nodes.node[:len(nodes.node)-1]
		}
	}
	//Update DB
}

func ProxyHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	output := ""
	body, _ := ioutil.ReadAll(r.Body)

	// Loop over all data nodes
	for i := 0; i < len(nodes.node); i++ {
		u := "http://" + nodes.node[i].address + "/update"
		reader := bytes.NewReader(body)

		req, err := http.NewRequest("POST", u, ioutil.NopCloser(reader))
		if err != nil {
			log.Fatal(err)
			fmt.Println(rw, "ERROR: Making request"+u)
		}
		req.Header = r.Header
		req.URL.Scheme = strings.Map(unicode.ToLower, req.URL.Scheme)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
			fmt.Println(rw, "ERROR: Sending request"+u)
		}
		output = output + u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n"
	}
	fmt.Println(output)
	fmt.Fprintf(rw, output)
}

func HandshakeHandler(rw http.ResponseWriter, r *http.Request) {
	handshake := mux.Vars(r)["handshake"]
	AddDataNode(handshake)

	fmt.Println(rw, "Handshake: " + handshake)
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
