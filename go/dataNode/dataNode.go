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

//string that points to the devise own home folder and port as subfolder
var basePath string = os.Getenv("HOME") + "/" + os.Getenv("PORT")

// Address to router
var routerAddress string = "localhost:9090"

func main() {
	r := mux.NewRouter()

	update := r.Path("/files").Subrouter()
	update.Methods("POST").HandlerFunc(FileUploadHandler)

	remove := r.Path("/deletefile/{id}").Subrouter()
	remove.Methods("DELETE").HandlerFunc(FileDeleteHandler)

	get := r.Path("/files/{id}").Subrouter()
	get.Methods("GET").HandlerFunc(FileGetHandler)

	info := r.Path("/files").Subrouter()
	info.Methods("GET").HandlerFunc(ListFilesHandler)

	// Delete all local files (if this is a crashed node in recovery)
	DeleteLocalFiles()

	// Notify master about our existence
	NotifyMaster()

	// Copy data from a sister data node
	GetListFromSister()

	http.ListenAndServe(":"+os.Getenv("PORT"), r)

}

//Delete local files
func DeleteLocalFiles() {
	fmt.Print("Clearing local files.. \n")

	//Delete folder an dall files in it
	os.RemoveAll(basePath)
	//Create new empty folder
	os.Mkdir(basePath, 0777)
}

// Handler to get file
func FileGetHandler(rw http.ResponseWriter, r *http.Request) {
	//Takes the id from url
	id := mux.Vars(r)["id"]

	//Set file path
	filePath := path.Join(basePath, id)

	//Reads file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	//Send status ok in response header
	rw.WriteHeader(http.StatusOK)
	//Send file in response body
	rw.Write(data)

	fmt.Printf("Data node: %s\n sent file: %s\n", os.Getenv("PORT"), id)
	fmt.Println(data)

}

//Handler to delete file
func FileDeleteHandler(rw http.ResponseWriter, r *http.Request) {
	//Gets id from url
	id := mux.Vars(r)["id"]

	//Set file path
	filePath := path.Join(basePath, id)

	//Removes file
	err := os.Remove(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	//Send status ok in response header
	rw.WriteHeader(http.StatusOK)

	fmt.Printf("Deleted file with id: %s\n", id)
}

//Handler to upload file
func FileUploadHandler(rw http.ResponseWriter, r *http.Request) {
	//Gets id from form
	id := r.FormValue("id")

	//Gets file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	//Set file path
	filePath := path.Join(basePath, id)

	defer file.Close()

	//Create the file
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(rw, "Unable to create the file for writing.")
		return
	}
	defer out.Close()

	//Write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(rw, err)
	}

	fmt.Printf("File uploaded successfully: %s\n", id)
}

//Send GET request to router to get address to primary master
func GetMasterAddress() string {
	//Set url
	url := "http://" + routerAddress + "/master"

	//Send request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	//Read response body
	body, err := ioutil.ReadAll(resp.Body)
	b := bytes.NewBuffer(body)

	return b.String()
}

//Send POST request to primary master that data node is up and running
func NotifyMaster() {
	//Get address to primary master
	masterAddress := GetMasterAddress()
	//Get own address
	nodeAddress := "localhost:" + os.Getenv("PORT")
	//Create url
	url := "http://" + masterAddress + "/handshake/" + nodeAddress
	//Create POST request
	r, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Making request" + url)
	}

	//Send request
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Sending request" + url)
	}

	fmt.Println(resp.Status + ": data node says hello")

}

//Function to get all files from another data node
func GetListFromSister() {
	//Get address to sister node
	sister := GetDataNodeAddress()
	//If sister is empty no copying will be initiated
	if sister == "" {
		fmt.Printf("No sister available\n")
		return
	}

	fmt.Printf("Sister: %s\n", sister)

	//Create url to get list of files
	url := "http://" + sister + "/files"

	//Send request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Sending request" + url)
	}
	defer resp.Body.Close()

	//Read response body
	body, err := ioutil.ReadAll(resp.Body)
	b := bytes.NewBuffer(body)

	//Call method to copy files
	CopySister(b.String(), sister)
}

//Function to contact master to get an ip to another data node
func GetDataNodeAddress() string {
	//Get address to primary master
	masterAddress := GetMasterAddress()
	//Create url to request
	url := "http://" + masterAddress + "/node"
	//Send request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	//Read response body
	body, err := ioutil.ReadAll(resp.Body)
	b := bytes.NewBuffer(body)

	//If the node address is the same as the node who sent the request
	//an empty string will be returned
	if b.String() == "localhost:"+os.Getenv("PORT") {
		return ""
	}

	return b.String()
}

//List all the files in home folder
func ListFiles() string {
	var allFiles string
	//Read file names in home folder
	files, _ := ioutil.ReadDir(basePath + "/")
	//Loop through files and writes the names to string
	for _, f := range files {
		allFiles += ("," + f.Name())
	}

	return allFiles
}

//Handler to GET list of files
func ListFilesHandler(rw http.ResponseWriter, r *http.Request) {
	//String of file names
	files := ListFiles()
	//Send status ok in response header
	rw.WriteHeader(http.StatusOK)
	//Send string in response body
	rw.Write([]byte(files))

}

//Send GET request for each file in file string
func CopySister(files string, sister string) {
	//Split string of file into array
	s := strings.Split(files, ",")

	//Remove first element in array since it is empty
	s[0] = s[len(s)-1]
	s = s[:len(s)-1]

	//Loop through array
	for _, id := range s {
		//Create url
		url := "http://" + sister + "/get" + id

		//Send request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		//Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		file := bytes.NewReader(body)

		//Create file path
		filePath := path.Join(basePath, id)

		//Create file
		out, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Unable to create the file for writing.")
			return
		}
		defer out.Close()

		//Write the content from GET to the file
		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("File uploaded successfully: %s\n", id)

	}
}
