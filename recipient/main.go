package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
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

func GetValidByte(src []byte) []byte {
	var str_buf []byte
	for _, v := range src {
		if v != 0 {
			str_buf = append(str_buf, v)
		}
	}
	return str_buf
}

func main() {
	upgrader := &websocket.Upgrader{}
	http.HandleFunc("/uploads", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		if r.URL.Query().Get("token") != config.Token {
			c.Close()
		}
		defer c.Close()
		for {
			mtype, msg, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			filename := string(GetValidByte(msg[:256]))

			for {
				dst, err := os.Create(path.Join("./backup/", filename))
				if err != nil {
					fmt.Println(err)
					namepath := strings.Split(filename, "/")
					os.MkdirAll(path.Join("./backup/", strings.Join(namepath[:len(namepath)-1], "/")), os.ModePerm)
					continue
				}
				defer dst.Close()
				io.Copy(dst, bytes.NewReader(msg[256:]))
				break
			}

			err = c.WriteMessage(mtype, []byte("next"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})
	log.Printf("Starting Server: 0.0.0.0:%d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
