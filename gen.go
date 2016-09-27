package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/beego/goyaml2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const (
	InsufficientPathError = "chemin insuffisant ou trop long pour trouver un cv.yaml"
)

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

func readFromGithub(filePath string) (string, error) {
	splitted := strings.Split(filePath, "/")
	var fileName string
	var fileSubPath string
	if l := len(splitted); l == 3 {
		fileName = splitted[2]
		fileSubPath = splitted[0] + "/" + splitted[1]
	} else {
		fileName = "cv.html"
		if l == 2 {
			fileSubPath = splitted[0] + "/" + splitted[1]
		} else if l == 1 {
			fileSubPath = splitted[0] + "/cv"
		} else {
			return "", InsufficientPathError
		}
	}
	addr := "https://raw.githubusercontent/" + fileSubPath + "/master/" + fileName

	return readFromUrl(addr)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	splitted := strings.Split(path, "/")
	var cvPath string
	if l := len(splitted); l > 2 {
		cvPath = splitted[0] + "/" + splitted[1] + "/cv.yaml"
	} else if l == 1 {
		cvPath = splitted[0] + "/cv/cv.yaml"
	} else {
		http.Error(w, InsufficientPathError, http.StatusInternalServerError)
		return
	}

	body, err := readFromGithub(cvPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplBody, err := readFromGithub(path)
	if err != nil {
		tmplBody, err2 := readFromGithub("dvaumoron")
		if err2 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	tmpl.Execute(w, obj)
}

func makeHandler(prefixe string) http.HandlerFunc {

}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/css/", makeHandler("css"))
	http.HandleFunc("/img/", makeHandler("img"))
	http.HandleFunc("/js/", makeHandler("js"))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
