package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/mux"
	//"log"
	"database/sql"
	//"database/sql/driver"
	//"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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

var mutex = &sync.Mutex{}

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

	update := r.Path("/files")
	update.Methods("POST").HandlerFunc(ProxyHandlerFunc)

	handshake := r.Path("/handshake/{nodeAddress}")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)

	deleteFile := r.Path("/deletefile/{id}")
	deleteFile.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	//Temp to test get method
	getFile := r.Path("/files/{id}")
	getFile.Methods("GET").HandlerFunc(GetFileHandler)

	getMasterIp := r.Path("/master/{ip}")
	getMasterIp.Methods("GET").HandlerFunc(AddMaster)

	shareNodes := r.Path("/share_nodes")
	shareNodes.Methods("GET").HandlerFunc(ShareNodes)

	getNodeIp := r.Path("/node/{ip}")
	getNodeIp.Methods("GET").HandlerFunc(GetNewNode)

	getSisterNode := r.Path("/node")
	getSisterNode.Methods("GET").HandlerFunc(GetSisterNode)

	getFilenames := r.Path("/get_filenames")
	getFilenames.Methods("GET").HandlerFunc(GetFilenames)

	NotifyRouter()

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}

// Gets all filenames from datanode
func GetFilenames(rw http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://" + nodes.node[0].address + "/files")
	if err != nil {
		fmt.Println("ERROR: Filenames")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		rw.Write(body)
	}
}

func GetFileHandler(rw http.ResponseWriter, r *http.Request) {
	if len(nodes.node) == 0 {
		fmt.Println(rw, "ERROR: No registered data nodes")
		return
	}

	//Connect to DB
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	if err != nil {
		fmt.Println("ERROR: Opening DB")
	}

	id := mux.Vars(r)["id"]

	//Get dataNode address from DB
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

	ip := nodes.node[rand.Intn(len(nodes.node))].address

	u := "http://" + ip + "/files/" + id //Specific url for every node

	req, err := http.NewRequest("GET", u, nil) //Create new request
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

	fmt.Println("Master sent GET req")

	fmt.Println("Master recieved file: ")
	fmt.Println(resp.Status)
}

func GetSisterNode(rw http.ResponseWriter, r *http.Request) {
	if len(nodes.node) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Not found"))
		return
	}
	//fix so that it's not the node asked that gets returned but her two sisternodes
	sister := nodes.node[0].address
	rw.Write([]byte(sister))
}

func ProxyHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	output := ""
	//Read body
	body, _ := ioutil.ReadAll(r.Body)

	// Loop over all data nodes
	for i := 0; i < len(nodes.node); i++ {
		u := "http://" + nodes.node[i].address + "/files"
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
}

//Adderar fil med name, year, course och faculty från input
func AddFile(file, year, course, faculty string) (string, string) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	result, err := db.Exec("INSERT INTO files (faculty, course, year, name) VALUES (?, ?, ?, ?)", faculty, course, year, file) //addera fil
	checkError(err)

	_, err = result.RowsAffected()
	if err != nil { //om inga rader blev affectade av insättning
		return file, "\nFILE: " + file + " COULD NOT BE ADDED, UNKNOWN ERROR" //något gick fel...
	} else {
		return file, "\nFILE ADDED: " + file //ADDERAD!
	}
}

//Adderar existerande fil till existerande nod
func AddFileToNode(serverIP, file string) string {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	server, err := db.Exec("SELECT id FROM servers WHERE ip=?", serverIP) //addera fil
	checkError(err)

	result, err := db.Exec("INSERT INTO fileserver VALUES (?, ?)", server, file) //addera fil
	checkError(err)

	_, err = result.RowsAffected()
	if err != nil { //om inga rader blev affectade av insättning
		return "\nFILE COULD NOT BE ADDED TO " + serverIP + ", UNKNOWN ERROR" //något gick fel...
	} else {
		return "\nFILE ADDED TO " + serverIP //ADDERAD!
	}
}

func getLastInsertFile(file string) string {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	var id string
	err = db.QueryRow("SELECT MAX(id) FROM files WHERE name=?", file).Scan(&id) //addera fil
	checkError(err)

	if err != sql.ErrNoRows {
		return id
	} else {
		return ""
	}
}

//Func for multicasting id of file to delete to nodes
func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"] //Get file id from request path

	// Loop over all data nodes
	for i := 0; i < len(nodes.node); i++ {
		u := "http://" + nodes.node[i].address + "/deletefile/" + id //Specific url for every node

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
		fmt.Printf("Master sen delete req to: %s\n", nodes.node[i].address)
		fmt.Println(resp.Status)
	}

	DeleteFileFromDB(id)
	fmt.Println("All files deleted")

}

func DeleteFileFromDB(id string) string {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	err = db.QueryRow("SELECT * FROM files WHERE id = ?", id).Scan(&id) //kolla om någon rad har fil id
	checkError(err)

	if err != sql.ErrNoRows { //om det kom tillbaka en rad
		result, err := db.Exec("DELETE FROM files WHERE id = ?", id) //ta bort fil
		checkError(err)
		_, err = result.RowsAffected()
		if err != nil { //om inga rader blev affectade av borttagningen
			return "\nFILE: " + id + " COULD NOT BE DELETED, UNKNOWN ERROR" //något gick fel...
		} else {
			return "\nFILE DELETED: " + id //BORTTAGEN!
		}
	} else {
		return "\nFILE: " + id + " COULD NOT BE DELETED, DO NOT EXIST" //vi kan ju inte ta bort något som inte finns...
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err, "<----ERROR----\n")
	}
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

//Adds datanode to list
func GetNewNode(rw http.ResponseWriter, r *http.Request) {
	ip := mux.Vars(r)["ip"]
	AddDataNode(ip)
}

//Returns all listed datanodes
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

//Gets all registered masters from other master
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

// Makes handshake with router
func NotifyRouter() {
	masterAddress := "localhost:" + os.Getenv("PORT") //Sets port to listen to

	//Connect to router
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
	//Get all registered masters
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
	GetNodes()//Get nodes from other master
	go MasterHeartbeat()//Start heartbeats
}

//Adding a dataNode to master list and DB
func AddDataNode(ip string) {
	node := node{address: ip, ok: true}
	mutex.Lock()
	ok := true
	for i := 0; i < len(nodes.node); i++ {
		if ip == nodes.node[i].address {
			ok = false
		}
	}
	if ok {
		nodes.node = append(nodes.node, node)
		fmt.Println("Added node: " + ip)
		fmt.Println("Number of nodes: " + string(len(nodes.node)))
		AddNodeToDB(ip)
	}
	mutex.Unlock()
}

func AddNodeToDB(ip string) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //kolla om någon rad redan har ip-numret
	checkError(err)

	if err == sql.ErrNoRows { //om det inte kom tillbaka några rader
		result, err := db.Exec("INSERT INTO servers (ip) VALUES (?)", ip) //addera server
		checkError(err)
		affected, err := result.RowsAffected()
		if err != nil { //om inga rader blev affectade av insättning
			fmt.Println("IP :%s COULD NOT BE ADDED, UNKNOWN ERROR", ip) //något gick fel...
		} else {
			fmt.Println("IP ADDED: %s AT ROW %s", ip, affected) //ADDERAD!
		}
	} else {
		fmt.Println("IP :%s COULD NOT BE ADDED, ALREADY EXIST", ip) //då adderar vi inte
	}
}

func RemoveDataNode(ip string) {
	//Remove node from master list
	mutex.Lock()
	if len(nodes.node) == 0 {
		return
	}
	for i := 0; i < len(nodes.node); i++ {
		if nodes.node[i].address == ip {
			nodes.node[i] = nodes.node[len(nodes.node)-1]
			nodes.node = nodes.node[:len(nodes.node)-1]
			fmt.Println("Removed node: " + ip)
			fmt.Println("Number of nodes: " + string(len(nodes.node)))
		}
	}
	mutex.Unlock()
	//Update DB
	DeleteNodeFromDB(ip)
}

//Tar bort noden med ip (input)
func DeleteNodeFromDB(ip string) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //kolla om någon rad har ip-numret
	checkError(err)

	if err != sql.ErrNoRows { //om det kom tillbaka en rad
		result, err := db.Exec("DELETE FROM servers WHERE ip = ?", ip) //ta bort server
		checkError(err)
		affected, err := result.RowsAffected()
		if err != nil { //om inga rader blev affectade av borttagningen
			fmt.Println("IP :%s COULD NOT BE DELETED, UNKNOWN ERROR", ip) //något gick fel...
		} else {
			fmt.Println("IP DELETED: %s AT ROW %s", ip, affected) //BORTTAGEN!
		}
	} else {
		fmt.Println("IP :%s COULD NOT BE DELETED, DO NOT EXIST", ip) //vi kan ju inte ta bort något som inte finns...
	}
}

func AddMaster(rw http.ResponseWriter, r *http.Request) {
	ip := mux.Vars(r)["ip"] //Get master ip
	AddMasterToList(ip)
}

func AddMasterToList(ip string) {
	ok := true
	for i := 0; i < len(mastersIp); i++ {
		if ip == mastersIp[i] {
			ok = false
		}
	}
	if ok {
		mastersIp = append(mastersIp, ip)
		fmt.Println("Registered new master: " + ip)
	}
}

func RemoveMaster(ip string) {
	//Remove node from master list
	if len(mastersIp) == 0 {
		return
	}
	for i := 0; i < len(mastersIp); i++ {
		if mastersIp[i] == ip {
			url := "http://" + routerAddress + "/master/" + mastersIp[i]
			r, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				fmt.Println("ERROR: Making request" + url)
			}
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				fmt.Println("ERROR: Sending request" + url)
			}
			fmt.Println("Removed: " + ip + " Router: " + resp.Status)
			mastersIp[i] = mastersIp[len(mastersIp)-1]
			mastersIp = mastersIp[:len(mastersIp)-1]
		}
	}
}

//Heartbeats. Checks masters, datanodes and router connections
func MasterHeartbeat() {
	heart := true
	for heart == true {
		routerOk := true
		time.Sleep(5000 * time.Millisecond)
		if len(mastersIp) > 0 {
			for i := 0; i < len(mastersIp); i++ {
				conn, err := net.DialTimeout("tcp", mastersIp[i], 10000*time.Millisecond)
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
				conn, err := net.DialTimeout("tcp", ip, 10000*time.Millisecond)
				if err != nil {
					fmt.Println("Timeout datanode: " + ip)
					RemoveDataNode(ip)
				} else {
					fmt.Println("Response datanode: " + conn.RemoteAddr().String() + " Status: OK")
				}
			}
		}
		conn, err := net.DialTimeout("tcp", routerAddress, 10000*time.Millisecond)
		if err != nil {
			routerOk = false
			fmt.Println("Timeout router: " + routerAddress)
			mastersIp = nil
			fmt.Println("Removed listed master...")
		} else {
			fmt.Println("Response router: " + conn.RemoteAddr().String() + " Status: OK")
		}
		for routerOk == false {
			fmt.Println("Retrying connection to router...")
			time.Sleep(5000 * time.Millisecond)
			conn, err := net.DialTimeout("tcp", routerAddress, 10000*time.Millisecond)
			if err == nil {
				fmt.Println("Reconnecting to " + conn.RemoteAddr().String() + " Status: OK")
				NotifyRouter()
				routerOk = true
				heart = false
			}
		}
	}
}

type File struct {
	Faculty string `json:"faculty"`
	Course  string `json:"course"`
	Year    string `json:"year"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

//json of all files
func getFilesAndFolders() string {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	rows, err := db.Query("SELECT * FROM files")
	checkError(err)

	var all_files string

	for rows.Next() {
		file := new(File)

		err = rows.Scan(&file.ID, &file.Faculty, &file.Course, &file.Year, &file.Name)

		checkError(err)

		all_files += file.ID + "," + file.Name + "|"
	}
	all_files = strings.TrimRight(all_files, "|")
	return all_files
}

func emptyDB() {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err)

	_, err = db.Exec("TRUNCATE TABLE files")
	checkError(err)

	_, err = db.Exec("TRUNCATE TABLE servers")
	checkError(err)
}
