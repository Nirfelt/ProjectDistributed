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
	get := r.Path("/{faculty}/{course}/{year}/{id}").Subrouter()
	get.Methods("GET").HandlerFunc(GetServerIdHoldingFile)

	add := r.Path("/add/{ip}").Subrouter()
	add.Methods("POST").HandlerFunc(AddNode)

	remove := r.Path("/delete/{ip}").Subrouter()
	remove.Methods("DELETE").HandlerFunc(DeleteNode)

	http.ListenAndServe(":8080", r)
}

func GetServerIdHoldingFile(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	checkError(err, rw)

	id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT ip FROM servers JOIN fileserver ON servers.id=fileserver.server_id WHERE file_id = ?", id)
	checkError(err, rw)

	for rows.Next() {
		var ip string

		err = rows.Scan(&ip)
		checkError(err, rw)

		fmt.Print("IP: ", ip)
		fmt.Fprintf(rw, "\nIP: %s", ip)
	}
	//return ip
}

func checkError(err error, rw http.ResponseWriter) {
	if err != nil {
		fmt.Fprintf(rw, "Error: ", err, "<----ERROR----\n")
	}
}

func AddNode(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")
	checkError(err, rw)

	ip := mux.Vars(r)["ip"]

	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //kolla om någon rad redan har ip-numret

	if err == sql.ErrNoRows { //om det inte kom tillbaka några rader
		result, err := db.Exec("INSERT INTO servers (ip) VALUES (?)", ip) //addera server
		checkError(err, rw)
		affected, err := result.RowsAffected()
		if err != nil { //om inga rader blev affectade av insättning
			fmt.Fprintf(rw, "\nIP :%s COULD NOT BE ADDED, UNKNOWN ERROR", ip) //något gick fel...
		} else {
			fmt.Fprintf(rw, "\nIP ADDED: %s AT ROW %s", ip, affected) //ADDERAD!
		}
	} else {
		fmt.Fprintf(rw, "\nIP :%s COULD NOT BE ADDED, ALREADY EXIST", ip) //då adderar vi inte
	}
}

func DeleteNode(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	ip := mux.Vars(r)["ip"]

	var id int
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //kolla om någon rad redan har ip-numret

	if err != sql.ErrNoRows { //om det kom tillbaka en rad
		result, err := db.Exec("DELETE FROM servers WHERE ip = ?", ip) //ta bort server
		checkError(err, rw)
		affected, err := result.RowsAffected()
		if err != nil { //om inga rader blev affectade av borttagningen
			fmt.Fprintf(rw, "\nIP :%s COULD NOT BE DELETED, UNKNOWN ERROR", ip) //något gick fel...
		} else {
			fmt.Fprintf(rw, "\nIP DELETED: %s AT ROW %s", ip, affected) //BORTTAGEN!
		}
	} else {
		fmt.Fprintf(rw, "\nIP :%s COULD NOT BE DELETED, DO NOT EXIST", ip) //vi kan ju inte ta bort något som inte finns...
	}
}

//func getJsonFilesAndFolders() {
//	//func get json of all files and folders
//}

//func AddFile(fileName, ip string) {
//	//func add file, file location ip
//}

//func DeleteFile(fileName, ip string) {
//	//func delete file, file location
//}
