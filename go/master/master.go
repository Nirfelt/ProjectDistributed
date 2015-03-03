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
	db, err := sql.Open("mysql", "ad0163:cwtfnrW5@195.178.235.60/websql/ad0163")

	checkError(err, rw)

	id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT ip FROM servers JOIN fileServer on servers.id=fileserver.server_id WHERE id=?", id)
	checkError(err, rw)

	for rows.Next() {
		var ip int

		err = rows.Scan(&ip)
		checkError(err, rw)

		fmt.Fprintf(rw, "IP: %s", ip)
	}
}

func checkError(err error, rw http.ResponseWriter) {
	if err != nil {
		fmt.Fprintf(rw, "Error: %s", err)
	}
}
