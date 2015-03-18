package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

var routerAddress string = "localhost:9090"

func main() {
	//Declare functions
	flag.Parse()

	r := mux.NewRouter()
	update := r.Path("/update")
	update.Methods("POST").HandlerFunc(ProxyHandlerFunc)
	handshake := r.Path("/handshake/{nodeAddress}")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)

	NotifyRouter()

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
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
	handshake := mux.Vars(r)["nodeAddress"]
	AddDataNode(handshake)

	fmt.Println("Handshake: " + handshake)
}

func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	faculty := mux.Vars(r)["faculty"]
	course := mux.Vars(r)["course"]
	year := mux.Vars(r)["year"]
	id := mux.Vars(r)["id"]

	fmt.Fprintf(rw, "Deleted file with id: %s, faculty: %s, course: %s, year: %s", id, faculty, course, year)
}

func NotifyRouter() {
	masterAddress := "localhost:" + os.Getenv("PORT")

	url := "http://" + routerAddress + "/handshake/" + masterAddress
	r, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Making request" + url)
	}

	//r.Body(nodeAddress)

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Sending request" + url)
	}
	output := url + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n"

	fmt.Println(output)
}

//func get datanode ip

//func return all files and folders
