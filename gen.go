package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"github.com/beego/goyaml2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type PathError struct {
	message string
}

func (err PathError) Error() string {
	return err.message
}

const InsufficientPathError = PathError{"path insufficient or too long for finding the file."}

func readFromUrl(addr string) (string, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := http.Client{Transport: transport}

	resp, err := client.Get(addr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func readFromGithub(fileSubPath, fileName string) (string, error) {
	return readFromUrl("https://raw.githubusercontent/" + fileSubPath + "/master/" + fileName)
}

func splitPath(filePath string, strict bool) (string, string, error) {
	splitted := strings.Split(filePath, "/")
	var fileName string
	var fileSubPath string
	if l := len(splitted); l == 3 {
		fileName = splitted[2]
		fileSubPath = splitted[0] + "/" + splitted[1]
	} else if strict {
		return "", "", InsufficientPathError
	} else {
		fileName = "cv.html"
		if l == 2 {
			fileSubPath = splitted[0] + "/" + splitted[1]
		} else if l == 1 {
			fileSubPath = splitted[0] + "/cv"
		} else {
			return "", "", InsufficientPathError
		}
	}
	return fileSubPath, fileName, nil
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fileSubPath, fileName, err := splitPath(r.URL.Path, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := readFromGithub(fileSubPath, "cv.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmplBody, err := readFromGithub(fileSubPath, fileName)
	if err != nil {
		tmplBody, err2 := readFromGithub("dvaumoron/cv", "cv.html")
		if err2 != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	tmpl, err := template.New("tmpl").parse(tmplBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj, err := goyaml2.Read(bytes.NewReader([]byte(body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	fileSubPath, fileName, err := splitPath(r.URL.Path[8:], true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := readFromGithub(fileSubPath, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write([]byte(body))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
        	log.Fatal("$PORT must be set")
    	}

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/static/", staticHandler)

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Println(err)
	}
}
