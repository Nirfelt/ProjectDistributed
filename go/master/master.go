package main

import (
	"fmt"
	//"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"database/sql"

	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := mux.NewRouter()
	get_server := r.Path("/get_server/{id}").Subrouter()
	get_server.Methods("GET").HandlerFunc(GetServerIdHoldingFile)

	get_filename := r.Path("/get_filename/{id}").Subrouter()
	get_filename.Methods("GET").HandlerFunc(GetFileName)

	add_server := r.Path("/add_server/{ip}").Subrouter()
	add_server.Methods("PUT").HandlerFunc(AddNode)

	delete_server := r.Path("/delete_server/{ip}").Subrouter()
	delete_server.Methods("DELETE").HandlerFunc(DeleteNode)

	add_file := r.Path("/add_file/{faculty}/{course}/{year}/{name}").Subrouter()
	add_file.Methods("PUT").HandlerFunc(AddFile)

	delete_file := r.Path("/delete_file/{id}").Subrouter()
	delete_file.Methods("DELETE").HandlerFunc(DeleteFile)

	http.ListenAndServe(":8080", r)
}

func GetServerIdHoldingFile(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	checkError(err, rw)

	id := mux.Vars(r)["id"]

	rows, err := db.Query("SELECT ip FROM servers JOIN fileserver ON servers.id=fileserver.server_id WHERE file_id = ?", id)
	checkError(err, rw)

	var all_ip []byte

	for rows.Next() {
		var ip string

		err = rows.Scan(&ip)
		checkError(err, rw)

		jsonStr, _ := json.Marshal(ip)

		all_ip = append(all_ip, jsonStr...)
	}
	rw.Write(all_ip)
}

func checkError(err error, rw http.ResponseWriter) {
	if err != nil {
		fmt.Fprintf(rw, "Error: ", err, "<----ERROR----\n")
	}
}

func checkError2(err error) {
	if err != nil {
		fmt.Println("Error: ", err, "<----ERROR----\n")
	}
}

func GetFileName(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	var file File

	file.ID = mux.Vars(r)["id"]

	err = db.QueryRow("SELECT name, year, course, faculty FROM files WHERE id = ?", file.ID).Scan(&file.Name, &file.Year, &file.Course, &file.Faculty) //kolla om någon rad har id

	if err != sql.ErrNoRows { //om det kom tillbaka en rad
		jsonStr, _ := json.Marshal(file)
		rw.Write(jsonStr)
		getJsonFilesAndFolders()
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
	err = db.QueryRow("SELECT * FROM servers WHERE ip = ?", ip).Scan(&id) //kolla om någon rad har ip-numret

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

type File struct {
	Faculty string `json:"faculty"`
	Course  string `json:"course"`
	Year    string `json:"year"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

//WIP
func getJsonFilesAndFolders() {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	rows, err := db.Query("SELECT * FROM files")

	checkError2(err)

	var all_json []byte

	for rows.Next() {
		file := new(File)

		err = rows.Scan(&file.ID, &file.Faculty, &file.Course, &file.Year, &file.Name)

		checkError2(err)

		jsonString, _ := json.Marshal(file)

		all_json = append(all_json, jsonString...)
	}

	//JAVASCRIPT exempel.....
	// var keys = Object.keys(dataMap)

	// for _, key := range keys {
	//         var innerJson []byte
	//         innerJson["name"] = keys[key]
	//         var innerMap = dataMap[keys[key]]

	//         if innerMap instanceof Array {
	//             innerJson["size"] = innerMap[0]
	//         } else if innerMap instanceof Object {

	//             var child = processHirarchiachalData(innerMap)
	//             innerJson["children"] = child
	//         }
	//         json.push(innerJson)

	// }
}

func AddFile(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	name := mux.Vars(r)["name"]
	year := mux.Vars(r)["year"]
	course := mux.Vars(r)["course"]
	faculty := mux.Vars(r)["faculty"]

	result, err := db.Exec("INSERT INTO files (faculty, course, year, name) VALUES (?, ?, ?, ?)", faculty, course, year, name) //addera fil
	checkError(err, rw)
	affected, err := result.RowsAffected()
	if err != nil { //om inga rader blev affectade av insättning
		fmt.Fprintf(rw, "\nFILE :%s COULD NOT BE ADDED, UNKNOWN ERROR", name) //något gick fel...
	} else {
		fmt.Fprintf(rw, "\nFILE ADDED: %s AT ROW %s", name, affected) //ADDERAD!
	}

	//ATT GÖRA: ADDERA FIL TILL NOD!
}

func AddFileToNode() {

}

func DeleteFile(rw http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "misa:password@tcp(mahsql.sytes.net:3306)/misa")

	id := mux.Vars(r)["id"]

	err = db.QueryRow("SELECT * FROM files WHERE id = ?", id).Scan(&id) //kolla om någon rad har fil id

	if err != sql.ErrNoRows { //om det kom tillbaka en rad
		result, err := db.Exec("DELETE FROM files WHERE id = ?", id) //ta bort fil
		checkError(err, rw)
		affected, err := result.RowsAffected()
		if err != nil { //om inga rader blev affectade av borttagningen
			fmt.Fprintf(rw, "\n>FILE :%s COULD NOT BE DELETED, UNKNOWN ERROR", id) //något gick fel...
		} else {
			fmt.Fprintf(rw, "\nFILE DELETED: %s AT ROW %s", id, affected) //BORTTAGEN!
		}
	} else {
		fmt.Fprintf(rw, "\nFILE :%s COULD NOT BE DELETED, DO NOT EXIST", id) //vi kan ju inte ta bort något som inte finns...
	}
}
