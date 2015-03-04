package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

var basePath string = os.Getenv("HOME")

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In the upload handler")

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	filePath := path.Join(basePath, "testOfUpload")

	defer file.Close()

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

func main() {
	http.HandleFunc("/", uploadHandler)
	http.ListenAndServe(":9090", nil)
}