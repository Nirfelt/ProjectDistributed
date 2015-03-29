package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/mux"
)

type masterlist struct {
	master []master
}

type master struct {
	address string
}

type Context struct {
	Title  string
	Static string
	Files  string
}

var masters = masterlist{} // List with masters (struct)
var id = 0

const STATIC_URL string = "/static/"

func main() {
	var staticPath = flag.String("staticPath", "static/", "Path to static files")
	flag.Parse()

	r := mux.NewRouter()

	go r.HandleFunc("/", Index)
	go r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(*staticPath))))

	update := r.Path("/files")
	update.Methods("POST").HandlerFunc(UploadHandler)

	getPrimary := r.Path("/master")
	getPrimary.Methods("GET").HandlerFunc(GetPrimaryHandler)

	handshake := r.Path("/handshake/{masterAddress}")
	handshake.Methods("POST").HandlerFunc(HandshakeHandler)

	getfile := r.Path("/files")
	getfile.Methods("GET").HandlerFunc(GetFileHandler)

	deletefile := r.Path("/deletefile")
	deletefile.Methods("GET").HandlerFunc(DeleteFileHandler)

	removeMaster := r.Path("/master/{ip}")
	removeMaster.Methods("DELETE").HandlerFunc(RemoveMaster)

	go Heartbeat()

	http.ListenAndServe(":9090", r)

}

func Index(w http.ResponseWriter, req *http.Request) {
	context := Context{Title: "Files: "}
	render(w, "list", context)
}

func render(w http.ResponseWriter, tmpl string, context Context) {
	context.Static = STATIC_URL
	context.Files = WriteFiles()
	tmpl_list := []string{"templates/index.html",
		fmt.Sprintf("templates/%s.html", tmpl)}
	t, err := template.ParseFiles(tmpl_list...)
	if err != nil {
		fmt.Println("template parsing error: ", err)
	}
	err = t.Execute(w, context)
	if err != nil {
		fmt.Println("template executing error: ", err)
	}
}

func WriteFiles() string {
	output := ""
	resp, err := http.Get("http://" + masters.master[0].address + "/get_filenames")
	if err != nil {
		fmt.Println("ERROR: Getting filelist")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		files := strings.Split(strings.TrimLeft(string(body), ","), ",")
		for i := 0; i < len(files); i++ {
			output += fmt.Sprintf("%s", files[i])
		}
	}

	return output
}

func UploadHandler(rw http.ResponseWriter, r *http.Request) {
	if len(masters.master) == 0 {
		fmt.Println(rw, "ERROR: No registered masters")
		return
	}
	output := ""
	body, _ := ioutil.ReadAll(r.Body)
	u := "http://" + masters.master[0].address + "/files"
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
	//fmt.Fprintf(rw, output)
	context := Context{Title: "TEST!"}
	render(rw, "list", context)
}

func GetPrimaryHandler(rw http.ResponseWriter, r *http.Request) {
	master := []byte(masters.master[0].address)
	rw.Write(master)
}

func GetFileHandler(rw http.ResponseWriter, r *http.Request) {
	if len(masters.master) == 0 {
		fmt.Println(rw, "ERROR: No registered masters")
		return
	}
	id := r.FormValue("id")

	u := "http://" + masters.master[0].address + "/files/" + id

	req, err := http.NewRequest("GET", u, nil)
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

	fmt.Println("Router sent get file req")

	fmt.Println("Router recieved file: ")
	fmt.Println(resp.Status)

	context := Context{Title: "TEST!"}
	render(rw, "list", context)
}

func DeleteFileHandler(rw http.ResponseWriter, r *http.Request) {
	if len(masters.master) == 0 {
		fmt.Println(rw, "ERROR: No registered masters")
		return
	}
	id := r.FormValue("id")

	u := "http://" + masters.master[0].address + "/deletefile/" + id

	req, err := http.NewRequest("DELETE", u, nil)
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
	fmt.Println("Router sent delete req")
	fmt.Printf("Router recieved %s\n", resp.Status)
	context := Context{Title: "TEST!"}
	render(rw, "list", context)
}

func AddMaster(address string) {

	master := master{address: address}
	masters.master = append(masters.master, master)

}

func RemoveMaster(rw http.ResponseWriter, r *http.Request) {
	ip := mux.Vars(r)["ip"]
	if len(masters.master) == 0 {
		return
	}
	for i := 0; i < len(masters.master); i++ {
		if masters.master[i].address == ip {
			fmt.Println("Removed master: " + ip)
			masters.master[i] = masters.master[len(masters.master)-1]
			masters.master = masters.master[:len(masters.master)-1]
		}
	}
}

func HandshakeHandler(rw http.ResponseWriter, r *http.Request) {
	handshake := mux.Vars(r)["masterAddress"]
	output := "No masters to update\n"
	mastersIp := ""
	// Loop over all masters
	if len(masters.master) > 0 {
		output = ""
		for i := 0; i < len(masters.master); i++ {
			u := "http://" + masters.master[i].address + "/master/" + handshake
			//Make new request
			resp, err := http.Get(u)
			if err != nil {
				fmt.Println(rw, "ERROR: Making request "+u)
			}
			output = output + u + "\tStatus: " + resp.Status + "\n"
			mastersIp = mastersIp + "," + masters.master[i].address
		}
	}
	rw.Write([]byte(mastersIp))
	AddMaster(handshake)
	fmt.Println("Handshake: " + handshake)
	fmt.Println(output)
}

func Heartbeat() {
	for {
		time.Sleep(5000 * time.Millisecond)
		if len(masters.master) == 1 {
			ip := masters.master[0].address
			conn, err := net.DialTimeout("tcp", ip, 3000*time.Millisecond)
			if err != nil {
				fmt.Println("Timeout: " + ip)
				fmt.Println("Removed master: " + ip)
				masters.master[0] = masters.master[len(masters.master)-1]
				masters.master = masters.master[:len(masters.master)-1]
			} else {
				fmt.Println("Response: " + conn.RemoteAddr().String() + " Status: OK")
			}
		}
	}
}
