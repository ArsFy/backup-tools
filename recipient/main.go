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
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Token string `json:"token"`
	Path  string `json:"path"`
}{
	Host:  "0.0.0.0",
	Port:  26543,
	Token: "",
	Path:  "./backup/",
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

func writeFile(filename string, data []byte) {
	for {
		dst, err := os.Create(path.Join(config.Path, filename))
		if err != nil {
			namepath := strings.Split(filename, "/")
			os.MkdirAll(path.Join(config.Path, strings.Join(namepath[:len(namepath)-1], "/")), os.ModePerm)
			continue
		}
		defer dst.Close()
		io.Copy(dst, bytes.NewReader(data))
		break
	}
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
				if err.Error() != "websocket: close 1006 (abnormal closure): unexpected EOF" {
					log.Println(err)
				} else {
					log.Println(r.RemoteAddr, "WebSocket Close")
				}
				break
			}

			isWrite := string(GetValidByte(msg[0:1])) == "1"
			filename := string(GetValidByte(msg[1:257]))

			fmt.Println(isWrite, filename)

			if isWrite {
				writeFile(filename, msg[257:])
			} else {
				os.Remove(path.Join(config.Path, filename))
			}

			err = c.WriteMessage(mtype, []byte("next"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})
	log.Printf("Starting Server: %s:%d\n", config.Host, config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), nil))
}
