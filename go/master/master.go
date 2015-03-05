package main

import (
	"fmt"
	//"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := mux.NewRouter()
	file := r.Path("/{faculty}/{course}/{year}/{id}").Subrouter()
	file.Methods("GET").HandlerFunc(searchDB)
	//file.Methods("POST").HandlerFunc(FileCreateHandler)
	//file.Methods("DELETE").HandlerFunc(FileDeletehandler)

	http.ListenAndServe(":8080", r)
}

func searchDB(rw http.ResponseWriter, r *http.Request) {

	fmt.Print("INIT!\n")
	//db, err := sql.Open("mysql", "ad0163:cwtfnrW5@195.178.235.60/ad0163")
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	checkError(err, rw)

	id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT ip FROM servers JOIN fileserver on servers.id=fileserver.server_id WHERE file_id=?", id)
	checkError(err, rw)

	for rows.Next() {
		var ip string

		err = rows.Scan(&ip)
		checkError(err, rw)

		fmt.Print("IP: ", ip)
		fmt.Fprintf(rw, "\nIP: %s", ip)
	}
	//return ip as string
}

func checkError(err error, rw http.ResponseWriter) {
	if err != nil {
		fmt.Print("Error: ", err, "<----ERROR----\n")
	}
}

//func get json of all files and folders

//func add file, file location ip

//func delete file, file location

//func add new node

//func delete node
