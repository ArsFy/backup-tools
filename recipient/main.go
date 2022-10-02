package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

var config = struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
}{
	Port:  26543,
	Token: "",
}

func init() {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Config Err:", err)
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("Config Err:", err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") == config.Token {
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			name := r.URL.Query().Get("name")
			if name == "" {
				name = part.FileName()
			}

			for {
				dst, err := os.Create("./backup/" + name)
				if err != nil {
					namepath := strings.Split(name, "/")
					os.MkdirAll(path.Join("./backup/", strings.Join(namepath[:len(namepath)-1], "/")), os.ModePerm)
					continue
				}
				defer dst.Close()
				io.Copy(dst, part)
				break
			}
			fmt.Fprintf(w, `ok`)
		}
	} else {
		fmt.Println("TokenErr:", r.RemoteAddr, r.URL.Query().Get("token"))
		fmt.Fprintf(w, `token`)
	}
}

func main() {
	http.HandleFunc("/backup", uploadHandler)
	fmt.Printf("Starting Server: 0.0.0.0:%d\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
