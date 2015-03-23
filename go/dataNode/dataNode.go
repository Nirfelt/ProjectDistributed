package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

//string that points to the devise own home folder
var basePath string = os.Getenv("HOME") + "/" + os.Getenv("PORT")
var routerAddress string = "localhost:9090"
var sisterNode1 string
var sisterNode2 string

func main() {
	r := mux.NewRouter()

	update := r.Path("/update").Subrouter()
	update.Methods("POST").HandlerFunc(FileUploadHandler)

	remove := r.Path("/delete/{id}").Subrouter()
	remove.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	get := r.Path("/get/{id}").Subrouter()
	get.Methods("GET").HandlerFunc(FileGetHandler)

	getAll := r.Path("/getall").Subrouter()
	getAll.Methods("GET").HandlerFunc(SendAllFiles)

	NotifyMaster()

	http.ListenAndServe(":"+os.Getenv("PORT"), r)

}

func FileGetHandler(rw http.ResponseWriter, r *http.Request) {
	//Takes the id from url
	id := mux.Vars(r)["id"]

	//set file path
	filePath := path.Join(basePath, id)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
	//rw.Header().Set(key, value)
	rw.Write(data)

}

func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("data node ok")
	id := mux.Vars(r)["id"]
	//id := r.FormValue("id")

	filePath := path.Join(basePath, id)

	err := os.Remove(filePath)
	if err != nil {
		// Better error handling would be nice..
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
	log.Printf("Deleted file with id: %s\n", id)
}

func FileUploadHandler(rw http.ResponseWriter, r *http.Request) {
	//id := mux.Vars(r)["id"]
	id := r.FormValue("id")
	fmt.Println("Thanks for the request")

	//fmt.Fprintf(rw, "Faculty: %s, Course: %s, Year: %s, Id: %s\n", faculty, course, year, id)
	// the FormFile function takes in the POST input id file
	file, _, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	filePath := path.Join(basePath, id)

	defer file.Close()

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(rw, "Unable to create the file for writing.")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(rw, err)
	}
	fmt.Fprintf(rw, "File uploaded successfully: %s\n", id)
}

func OnStartUp() string {
	ListFiles()

	fmt.Println("Who is primary master?")

	url := "http://" + routerAddress + "/getprimary"

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	b := bytes.NewBuffer(body)

	return b.String()
}

func NotifyMaster() {
	masterAddress := OnStartUp()

	nodeAddress := "localhost:" + os.Getenv("PORT")
	url := "http://" + masterAddress + "/handshake/" + nodeAddress
	r, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Making request" + url)
	}

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Sending request" + url)
	}
	output := url + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto + "\n\n"

	fmt.Println(output)

	CopySister()

}

//Function to get all files from another data node
func CopySister() {
	//sisters := GetNodeIP(masterAddress)
	sisters := "8081,8082"
	s := strings.Split(sisters, ",")
	sisterNode1 = s[0]
	sisterNode2 = s[1]
	fmt.Println(sisterNode1)
	fmt.Println(sisterNode2)

	//contact sister 1
	//else contact sister 2
	//if file allready exists then overwrite it
	//let master know everything is ok
	//or send a GET for every file in own list, if not ok
	//then send a new GET to recieve that file.

}

//Add all files to list
func ListFiles() {

}

//Send files
func SendAllFiles(rw http.ResponseWriter, r *http.Request) {

}

//Function to contact master to get an ip to another data node
func GetNodeIP(masterIP string) string {
	url := "http://" + masterIP + "/getconnectednod/" + os.Getenv("PORT")

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	b := bytes.NewBuffer(body)

	return b.String()
}

//Add timestamp on datanodes

//Function to tell master when a file has been saved
