package main

import (
	"bytes"
	"crypto/tls"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type PathError string

func (err PathError) Error() string {
	return string(err)
}

const InsufficientPathError = PathError("path insufficient for finding the file.")

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
	return string(body), err
}

func readFromGithub(repoPath, folderPath, fileName string) (string, error) {
	var filePath string
	if folderPath == "" {
		filePath = fileName
	} else {
		filePath = folderPath + "/" + fileName
	}
	if local {
		body, err := ioutil.ReadFile(localPath + filePath)
		return string(body), err
	}
	return readFromUrl("https://raw.githubusercontent.com/" + repoPath + "/master/" + filePath)
}

func splitWithoutBlank(s string) []string {
	splitted := strings.Split(s, "/")
	res := make([]string, 0)
	for _, value := range splitted {
		if value != "" {
			res = append(res, value)
		}
	}
	return res
}

func splitPath(path string, strict bool) (string, string, string, error) {
	splitted := splitWithoutBlank(path)
	var repoPath, fileName string
	folderPath := ""
	if l := len(splitted); l > 3 {
		repoPath = splitted[0] + "/" + splitted[1]
		first := true
		for i := 2; i < l-1; i++ {
			if first {
				first = false
			} else {
				folderPath += "/"
			}
			folderPath += splitted[i]
		}
		fileName = splitted[l-1]
	} else if strict {
		return "", "", "", InsufficientPathError
	} else {
		fileName = "cv.html"
		if l == 2 {
			repoPath = splitted[0] + "/" + splitted[1]
		} else if l == 1 {
			repoPath = splitted[0] + "/cv"
		} else {
			return "", "", "", InsufficientPathError
		}
	}
	return repoPath, folderPath, fileName, nil
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	repoPath, folderPath, fileName, err := splitPath(r.URL.Path, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := readFromGithub(repoPath, folderPath, "cv.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmplBody, err := readFromGithub(repoPath, folderPath, fileName)
	if err != nil {
		var err2 error
		tmplBody, err2 = readFromGithub("dvaumoron/cv", "", "cv.html")
		if err2 != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	tmpl, err := template.New("tmpl").Parse(tmplBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj, err := yaml.Parse(bytes.NewReader([]byte(body)))
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
	repoPath, folderPath, fileName, err := splitPath(r.URL.Path[8:], true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := readFromGithub(repoPath, folderPath, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write([]byte(body))
}

var local = false
var localPath = ""

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	localPath = os.Getenv("LOCAL_PATH")
	if localPath != "" {
		local = true
	}

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/static/", staticHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println(err)
	}
}
