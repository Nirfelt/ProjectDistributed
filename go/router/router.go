package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"unicode"
)

type masterlist struct {
	master []master
}

type master struct {
	address string
}

var masters = masterlist{} // List with masters (struct)

func main() {
	r := mux.NewRouter()

	update := r.Path("/update")
	update.Methods("POST").HandlerFunc(UploadHandler)
	getPrimary := r.Path("/getprimary")
	getPrimary.Methods("GET").HandlerFunc(GetPrimaryHandler)
	handshake := r.Path("/handshake/{masterAddress}")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)
	getfile := r.Path("/getfile/{id}")
	getfile.Methods("GET").HandlerFunc(GetFileHandler)
	deletefile := r.Path("/deletefile/{id}")
	deletefile.Methods("DELETE").HandlerFunc(DeleteFileHandler)

	http.ListenAndServe(":9090", r)

}

func UploadHandler(rw http.ResponseWriter, r *http.Request) {
	if len(masters.master) == 0 {
		fmt.Println(rw, "ERROR: No registered masters")
		return
	}
	output := ""
	body, _ := ioutil.ReadAll(r.Body)
	u := "http://" + masters.master[0].address + "/update"
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
	output = u + "\nStatus: " + resp.Status + "\nProtocol: " + resp.Proto
	fmt.Println(output)
	fmt.Fprintf(rw, output)
}

func GetPrimaryHandler(rw http.ResponseWriter, r *http.Request) {
	master := []byte(masters.master[0].address)
	rw.Write(master)
}

func GetFileHandler(rw http.ResponseWriter, r *http.Request) {

}

func DeleteFileHandler(rw http.ResponseWriter, r *http.Request) {

}

func AddMaster(address string) {
	master := master{address: address}
	masters.master = append(masters.master, master)
}

func RemoveMaster(address string) {
	if len(masters.master) == 0 {
		return
	}
	for i := range masters.master {
		if masters.master[i].address == address {
			masters.master[i] = masters.master[len(masters.master)-1]
			masters.master = masters.master[:len(masters.master)-1]
		}
	}
}

func HandshakeHandler(rw http.ResponseWriter, r *http.Request) {
	handshake := mux.Vars(r)["masterAddress"]
	AddMaster(handshake)

	fmt.Println("Handshake: " + handshake)
}
