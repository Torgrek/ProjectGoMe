package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func httpModule() {

	InitRouter()

}

func InitRouter() {
	http.Handle("/", http.FileServer(http.Dir("./site/static")))
	initHandlers()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}

}

func initHandlers() {
	http.HandleFunc("/funcs/upload", uploadfile)
	http.HandleFunc("/catalog", catalog)
	http.HandleFunc("/all.min.js", SendScripts)
	http.HandleFunc("/bootstrap.min.js", SendBootstrap)
	http.HandleFunc("/bootstrap.min.js.map", SendBootstrapMap)
	http.HandleFunc("/scripts.js", SendFontAwesome)
	http.HandleFunc("/bootstrap.min.css", SendBootstrapCSS)
	http.HandleFunc("/bootstrap.min.css.map", SendBootstrapCSSMap)
}

func SendScripts(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/js/scripts.js")
}

func SendBootstrap(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/js/bootstrap/bootstrap.min.js")
}

func SendFontAwesome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/js/fontawesome/all.min.js")
}

func SendBootstrapCSS(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/css/bootstrap/bootstrap.min.css")
}

func SendBootstrapCSSMap(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/css/bootstrap/bootstrap.min.css.map")
}
func SendBootstrapMap(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "site/js/bootstrap/bootstrap.min.js.map")
}

func catalog(w http.ResponseWriter, r *http.Request) {

	dbDriver := globalruntimeparams.driver
	rows, err := dbDriver.Query("SELECT guildid, userid, id, filename, active, bought FROM catalog")
	checkIfNil(err)
	var data []catalogData

	for rows.Next() {
		currentRow := catalogData{}
		err := rows.Scan(&currentRow.Id, &currentRow.Guildid, &currentRow.Userid, &currentRow.Filename, &currentRow.Active, &currentRow.Bought)
		checkIfNil(err)
		data = append(data, currentRow)
	}

	tmpl, _ := template.ParseFiles("site/templates/catalog.html")

	if tmpl != nil {
		tmpl.Execute(w, data)
	}
}

type catalogData struct {
	Id       int
	Guildid  int64
	Userid   int64
	Filename string
	Active   bool
	Bought   bool
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

	extension := strings.ReplaceAll(filepath.Ext(handler.Filename), ".", "")
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
