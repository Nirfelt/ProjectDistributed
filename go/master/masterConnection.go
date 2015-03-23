package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	//"log"
	"net/http"
	"os"
	"strings"
	"unicode"
	"encoding/json"
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
var masterDB string = "localhost:9191"

func main() {
	//Declare functions
	flag.Parse()
	r := mux.NewRouter()

	update := r.Path("/update")
	update.Methods("POST").HandlerFunc(ProxyHandlerFunc)

	handshake := r.Path("/handshake/{nodeAddress}")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)

	deleteFile := r.Path("/delete/{id}")
	deleteFile.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	getFile := r.Path("/get_file/{id}")
	getFile.Methods("GET").HandlerFunc(GetFileHandler)

	NotifyRouter()
	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}


func GetFileHandler(rw http.ResponseWriter, r *http.Request) {
	//Get dataNode address from masterDB
	id := mux.Vars(r)["id"]
	resp, err := http.Get("http://" + masterDB + "/get_server/" + id)
	if err != nil {
		fmt.Println("ERROR: Sending request"+masterDB)
	}
	//Handle response
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var data []string
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("ERROR: Unmarshall json")
	}
	for i := 0; i < len(data); i++ {
		fmt.Println(data[i])
	}
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
			fmt.Println(rw, "ERROR: Making request"+u)
		}
		req.Header = r.Header
		req.URL.Scheme = strings.Map(unicode.ToLower, req.URL.Scheme)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(rw, "ERROR: Sending request"+u)
		}
		output = output + u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n"
	}
	fmt.Println(output)
	fmt.Fprintf(rw, output)
}

//Func for multicasting id of file to delete to nodes
func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("master ok")
	id := mux.Vars(r)["id"] //Get file id from request path
	output := ""

	// Loop over all data nodes
	for i := 0; i < len(nodes.node); i++ {
		u := "http://" + nodes.node[i].address + "/delete/" + id //Specific url for every node

		req, err := http.NewRequest("DELETE", u, nil) //Create new request
		if err != nil {
			fmt.Println(rw, "ERROR: Making request"+u)
		}
		req.Header = r.Header
		req.URL.Scheme = strings.Map(unicode.ToLower, req.URL.Scheme)

		client := &http.Client{}
		resp, err := client.Do(req) //Send request, get response
		if err != nil {
			fmt.Println(rw, "ERROR: Sending request"+u)
		}
		output = output + u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n" //Output string
	}
	fmt.Println(output)
	fmt.Fprintf(rw, output)
}

//Handles new datanodes connecting
func HandshakeHandler(rw http.ResponseWriter, r *http.Request) {
	handshake := mux.Vars(r)["nodeAddress"]
	AddDataNode(handshake)

	fmt.Println("Handshake: " + handshake)
}

func NotifyRouter() {
	masterAddress := "localhost:" + os.Getenv("PORT")

	url := "http://" + routerAddress + "/handshake/" + masterAddress
	r, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("ERROR: Making request" + url)
	}

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		fmt.Printf("ERROR: Sending request" + url)
	}
	output := url + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n"

	fmt.Println(output)
}

func AddDataNode(address string) {
	node := node{address: address, ok: true}
	nodes.node = append(nodes.node, node)
	//update DB
	u := "http://" + masterDB + "/add_server/" + address

	req, err := http.NewRequest("PUT", u, nil) //Create new request
	if err != nil {
		fmt.Println("ERROR: Making request"+u)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: Sending request"+u)
	}
	fmt.Println(u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n")
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
	u := "http://" + masterDB + "/delete_server/" + node

	req, err := http.NewRequest("DELETE", u, nil) //Create new request
	if err != nil {
		fmt.Println("ERROR: Making request"+u)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: Sending request"+u)
	}
	fmt.Println(u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n")
}

//func get datanode ip

//func return all files and folders
