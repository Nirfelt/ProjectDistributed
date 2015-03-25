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

	//Path to recieve Get string list of files /getfileinfo

	info := r.Path("/getfileinfo").Subrouter()
	info.Methods("GET").HandlerFunc(ListFilesHandler)

	// Delete all local files (if this is a crashed node in recovery)
	DeleteLocalFiles()

	// Notify master about our existence
	NotifyMaster()

	// Copy data from a sister data node
	GetListFromSister()

	http.ListenAndServe(":"+os.Getenv("PORT"), r)

}

func DeleteLocalFiles() {
	fmt.Print("Clearing local files.. ")

	os.RemoveAll(basePath)
	os.Mkdir(basePath, 0777)

	fmt.Println("OK")
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
	id := r.FormValue("id")
	fmt.Println("Thanks for the request")

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

func GetMasterAddress() string {
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
	masterAddress := GetMasterAddress()
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

}

//Function to get all files from another data node
func GetListFromSister() {
	sister := GetDataNodeAddress()
	if sister == "" {
		fmt.Printf("No sister available")
		return
	}

	fmt.Printf("Sis: %s\n", sister)

	url := "http://" + sister + "/getfileinfo"

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
		fmt.Printf("ERROR: Sending request" + url)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	b := bytes.NewBuffer(body)

	CopySister(b.String(), sister)
}

//Function to contact master to get an ip to another data node
func GetDataNodeAddress() string {
	masterAddress := GetMasterAddress()
	url := "http://" + masterAddress + "/sisternode"

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	b := bytes.NewBuffer(body)

	if b.String() == "localhost:"+os.Getenv("PORT") {
		return ""
	}
	return b.String()
}

func ListFiles() string {
	var allFiles string
	files, _ := ioutil.ReadDir(basePath + "/")
	for _, f := range files {
		allFiles += ("," + f.Name())
		fmt.Println(f.Name())

	}
	fmt.Println(allFiles)
	return allFiles
}

func ListFilesHandler(rw http.ResponseWriter, r *http.Request) {
	files := ListFiles()
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(files))

}

func CopySister(files string, sister string) {
	fmt.Println(files)
	s := strings.Split(files, ",")
	fmt.Println(s)
	//s = append(s(0), s(1)...)
	//s = s[0 : len(s)-1]
	s[0] = s[len(s)-1]
	s = s[:len(s)-1]

	for _, id := range s {
		url := "http://" + sister + "/get" + id

		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		file := bytes.NewReader(body)

		filePath := path.Join(basePath, id)

		out, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Unable to create the file for writing.")
			return
		}

		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("File uploaded successfully: %s\n", id)

	}
}
