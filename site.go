package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/lib/pq"
)

func httpModule() {

	InitRouter()

}

func InitRouter() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	initHandlers()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}

}

func initHandlers() {
	http.HandleFunc("/funcs/upload", uploadfile)
}

func uploadfile(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("SoundFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploading File: %+v\n", handler.Filename)

	extension := filepath.Ext(handler.Filename)
	isCorrectFileType := false

	for _, item := range globalruntimeparams.aviliablefiles {
		if extension == item {
			isCorrectFileType = true
			break
		}
	}

	if !isCorrectFileType {
		fmt.Printf("Wrong file type: %+v\n", handler.Filename)
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	_, err = globalruntimeparams.driver.Exec("INSERT INTO soundlib (filename, bytediv) values($1::text, $2::bytea)", handler.Filename, fileBytes)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Added " + handler.Filename)
	}
}
