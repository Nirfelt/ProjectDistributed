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
	"database/sql"
	//"database/sql/driver"
	"math/rand"
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
var masterDB string = "localhost:9191"
var mastersIp []string

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

	getMasterIp := r.Path("/master_ip/{ip}")
	getMasterIp.Methods("GET").HandlerFunc(AddMaster)

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
		fmt.Printf("ERROR: Sending request" + url + "\n")
	}
	fmt.Println("Handshake: " + routerAddress)

	body, _ := ioutil.ReadAll(resp.Body)
	ips := strings.Split(string(body), ",")
	if len(ips) > 0 {
		for i := 0; i < len(ips); i++ {
			if ips[i] != ""{
				AddMasterToList(ips[i])
			}
		}
	}
}

//Adding a dataNode to master list and DB
func AddDataNode(ip string) {
	node := node{address: ip, ok: true}
	nodes.node = append(nodes.node, node)
	//Connect to DB
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	if err != nil {
		fmt.Printf("ERROR: Open DB")
	}
	//Making DB insert
	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //check if ip exists

	if err == sql.ErrNoRows { //Check return rows
		result, err := db.Exec("INSERT INTO servers (ip) VALUES (?)", ip) //add server
		if err != nil {
			fmt.Println("\nIP :%s INSERT FAILED", ip)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			fmt.Println("\nIP :%s COULD NOT BE ADDED, UNKNOWN ERROR", ip)
		} else {
			fmt.Println("\nIP ADDED: %s AT ROW %s", ip, affected) //ADDED!
		}
	} else {
		fmt.Println("\nIP :%s COULD NOT BE ADDED, ALREADY EXIST", ip)
	}
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
		}
	}
	//Update DB
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa") //Open DB connection
	if err != nil {
		fmt.Println("ERROR: Open DB")
	}
	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //check if any row has the ip
	if err != sql.ErrNoRows { //If a row is returned
		result, err := db.Exec("DELETE FROM servers WHERE ip = ?", ip) //Remove server server
		if err != nil {
			fmt.Println("\nIP :%s DELETE FAILED", ip)
		}
		affected, err := result.RowsAffected()
		if err != nil { //If no rows were affected
			fmt.Println("\nIP :%s COULD NOT BE DELETED, UNKNOWN ERROR", ip)
		} else {
			fmt.Println("\nIP DELETED: %s AT ROW %s", ip, affected) //REMOVED!
		}
	} else {
		fmt.Println("\nIP :%s COULD NOT BE DELETED, DO NOT EXIST", ip) //vi kan ju inte ta bort något som inte finns...
	}
}

func AddMaster(rw http.ResponseWriter, r *http.Request){
	ip := mux.Vars(r)["ip"] //Get master ip
	AddMasterToList(ip)
}

func AddMasterToList(ip string){
	mastersIp = append(mastersIp, ip)
	fmt.Println("Registered new master: " + ip)
}

//func get datanode ip

//func return all files and folders
