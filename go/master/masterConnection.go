package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/mux"
	//"log"
	"database/sql"
	"net"
	"net/http"
	"os"
	"strings"
	"unicode"
	//"database/sql/driver"
	"math/rand"
	"time"
	//_ "github.com/go-sql-driver/mysql"
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

//var masterDB string = "localhost:9191"
var mastersIp []string

func main() {
	//Declare functions
	flag.Parse()
	r := mux.NewRouter()

	update := r.Path("/update")
	go update.Methods("POST").HandlerFunc(ProxyHandlerFunc)

	handshake := r.Path("/handshake/{nodeAddress}")
	go handshake.Methods("POST").HandlerFunc(HandshakeHandler)

	deleteFile := r.Path("/delete/{id}")
	go deleteFile.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	getFile := r.Path("/get_file/{id}")
	go getFile.Methods("GET").HandlerFunc(GetFileHandler)

	getMasterIp := r.Path("/master_ip/{ip}")
	go getMasterIp.Methods("GET").HandlerFunc(AddMaster)

	shareNodes := r.Path("/share_nodes")
	go shareNodes.Methods("GET").HandlerFunc(ShareNodes)

	getNodeIp := r.Path("/node/{ip}")
	go getNodeIp.Methods("GET").HandlerFunc(GetNewNode)

	getSisterNode := r.Path("/sisternode")
	go getSisterNode.Methods("GET").HandlerFunc(GetSisterNode)

	NotifyRouter()
	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}

func GetFileHandler(rw http.ResponseWriter, r *http.Request) {
	//Connect to DB
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	if err != nil {
		fmt.Println("ERROR: Opening DB")
	}
	//Get dataNode address from DB
	id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT ip FROM servers JOIN fileserver ON servers.id=fileserver.server_id WHERE file_id = ?", id)
	if err != nil {
		fmt.Println("ERROR: SQL statement DB")
	}
	//Loop all ip to a list
	var all_ip []string
	for rows.Next() {
		var ip string

		err = rows.Scan(&ip)
		if err != nil {
			fmt.Println("ERROR: row.Scan")
		}

		all_ip = append(all_ip, ip)
	}
	//Return random ip from list
	rw.Write([]byte(all_ip[rand.Intn(len(all_ip))]))
}

func GetSisterNode(rw http.ResponseWriter, r *http.Request) {
	if len(nodes.node) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Not found"))
		return
	}

	sister := nodes.node[0].address
	rw.Write([]byte(sister))
}

func ProxyHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	output := ""
	//Read body
	body, _ := ioutil.ReadAll(r.Body)

	// Loop over all data nodes
	for i := 0; i < len(nodes.node); i++ {
		u := "http://" + nodes.node[i].address + "/update"
		reader := bytes.NewReader(body)
		//Create new request
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
	//Add to DB
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
	fmt.Println("Handshake: " + handshake + ", datanode")
	AddDataNode(handshake)
	//Send to other masters
	for i := 0; i < len(mastersIp); i++ {
		resp, err := http.Get("http://" + mastersIp[i] + "/node/" + handshake)
		if err != nil {
			fmt.Println("ERROR: Making request to: " + mastersIp[i])
		} else {
			fmt.Println("Node: " + handshake + "\tSent to: " + mastersIp[i] + "\tStatus: " + resp.Status)
		}
	}
}

func GetNewNode(rw http.ResponseWriter, r *http.Request) {
	ip := mux.Vars(r)["ip"]
	AddDataNode(ip)
}

func ShareNodes(rw http.ResponseWriter, r *http.Request) {
	output := ""
	if len(nodes.node) > 0 {
		for i := 0; i < len(nodes.node); i++ {
			output = output + "," + nodes.node[i].address
			fmt.Println("Sent node: " + nodes.node[i].address)
		}
	}
	rw.Write([]byte(output))
}

func GetNodes() {
	if len(mastersIp) > 0 {
		url := "http://" + mastersIp[0] + "/share_nodes"
		r, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("ERROR: Making request " + url)
		}
		client := &http.Client{}
		resp, err := client.Do(r)
		if err != nil {
			fmt.Println("ERROR: Sending request " + url)
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			ips := strings.Split(string(body), ",")
			if len(ips) > 0 {
				for i := 0; i < len(ips); i++ {
					if ips[i] != "" {
						AddDataNode(ips[i])
					}
				}
			}
		}
	}
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
		fmt.Printf("ERROR: Sending request" + url + "\n")
	}
	fmt.Println("Handshake: " + routerAddress + ", router")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR: Recieving master ip")
	}
	ips := strings.Split(string(body), ",")
	if len(ips) > 0 {
		for i := 0; i < len(ips); i++ {
			if ips[i] != "" {
				AddMasterToList(ips[i])
			}
		}
	}
	GetNodes()
	go MasterHeartbeat()
}

//Adding a dataNode to master list and DB
func AddDataNode(ip string) {
	node := node{address: ip, ok: true}
	nodes.node = append(nodes.node, node)
	fmt.Println("Added node: " + ip)
	//Connect to DB

}

func RemoveDataNode(ip string) {
	//Remove node from master list
	if len(nodes.node) == 0 {
		return
	}
	for i := range nodes.node {
		if nodes.node[i].address == ip {
			nodes.node[i] = nodes.node[len(nodes.node)-1]
			nodes.node = nodes.node[:len(nodes.node)-1]
			fmt.Println("Removed node: " + ip)
		}
	}
	//Update DB
}

func AddMaster(rw http.ResponseWriter, r *http.Request) {
	ip := mux.Vars(r)["ip"] //Get master ip
	AddMasterToList(ip)
}

func AddMasterToList(ip string) {
	mastersIp = append(mastersIp, ip)
	fmt.Println("Registered new master: " + ip)
}

func RemoveMaster(ip string) {
	//Remove node from master list
	if len(mastersIp) == 0 {
		return
	}
	for i := 0; i < len(mastersIp); i++ {
		if mastersIp[i] == ip {
			url := "http://" + routerAddress + "/remove_master/" + mastersIp[i]
			r, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				fmt.Printf("ERROR: Making request" + url)
			}
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				fmt.Printf("ERROR: Sending request" + url + "\n")
			}
			fmt.Println("Removed: " + ip + " Router: " + resp.Status)
			mastersIp[i] = mastersIp[len(mastersIp)-1]
			mastersIp = mastersIp[:len(mastersIp)-1]
		}
	}
}

func MasterHeartbeat() {
	for {
		time.Sleep(5000 * time.Millisecond)
		if len(mastersIp) > 0 {
			for i := 0; i < len(mastersIp); i++ {
				conn, err := net.DialTimeout("tcp", mastersIp[i], 3000*time.Millisecond)
				if err != nil {
					fmt.Println("Timeout master: " + mastersIp[i])
					RemoveMaster(mastersIp[i])
				} else {
					fmt.Println("Response master: " + conn.RemoteAddr().String() + " Status: OK")
				}
			}
		}
		if len(nodes.node) > 0 {
			for i := 0; i < len(nodes.node); i++ {
				ip := nodes.node[i].address
				conn, err := net.DialTimeout("tcp", ip, 3000*time.Millisecond)
				if err != nil {
					fmt.Println("Timeout datanode: " + ip)
					RemoveDataNode(ip)
				} else {
					fmt.Println("Response datanode: " + conn.RemoteAddr().String() + " Status: OK")
				}
			}
		}
	}
}

func InitiateBully() {
	//Get masters from router
	//Send GET request to masters with higher id
	//
}

func BullyRespond() {
	//If ID is higher then the one asked reply that
	//Send out new InitiateBully
}

func NotifyElectionResult() {
	//Tell router that a new primary has been selected
}

//func return all files and folders
