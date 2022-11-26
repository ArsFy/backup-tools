package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/schollz/progressbar/v3"
)

var config struct {
	Server  string   `json:"server"`
	Token   string   `json:"token"`
	Path    string   `json:"path"`
	Exclude []string `json:"exclude"`
}

var cache []string
var c *websocket.Conn

func init() {
	// Config
	file, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Config Err:", err)
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("Config Err:", err)
	}
	// Cache
	file, err = os.ReadFile("./cache.json")
	if err != nil {
		fmt.Println("Config Err:", err)
	}
	err = json.Unmarshal(file, &cache)
	if err != nil {
		fmt.Println("Config Err:", err)
	}
	// WS
	c, _, err = websocket.DefaultDialer.Dial(config.Server+"/uploads?token="+url.QueryEscape(config.Token), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
}

// filter exist files
func Arrcmp(src []string, dest []string) []string {
	msrc := make(map[string]byte)
	mall := make(map[string]byte)
	var set []string
	for _, v := range src {
		msrc[v] = 0
		mall[v] = 0
	}
	for _, v := range dest {
		l := len(mall)
		mall[v] = 1
		if l == len(mall) {
			set = append(set, v)
		}
	}
	for _, v := range set {
		delete(mall, v)
	}
	var added []string
	for v := range mall {
		_, exist := msrc[v]
		if !exist {
			added = append(added, v)
		}
	}
	return added
}

// Is file exclude
func isEx(name string) bool {
	for _, j := range config.Exclude {
		if strings.Contains(name, j) {
			return true
		}
	}
	return false
}

// Get All Files
func GetAllFile(pathname string) ([]string, error) {
	result := []string{}

	fis, err := os.ReadDir(pathname)
	if err != nil {
		fmt.Println("ReadErr:", err)
		return result, err
	}

	for _, fi := range fis {
		fullname := path.Join(pathname, fi.Name())
		if isEx(fullname) {
			continue
		}
		if fi.IsDir() {
			temp, err := GetAllFile(fullname)
			if err != nil {
				fmt.Println("ReadErr:", err)
				return result, err
			}
			result = append(result, temp...)
		} else {
			result = append(result, strings.Replace(fullname, config.Path, "", 1))
		}
	}

	return result, nil
}

// Bytes Combine
func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer
	for index := 0; index < len(pBytes); index++ {
		buffer.Write(pBytes[index])
	}
	return buffer.Bytes()
}

// Send File
func flc(j string, bar *progressbar.ProgressBar) {
	var filename [256]byte
	for ii, jj := range j {
		filename[ii] = byte(jj)
	}

	file, err := os.ReadFile(config.Path + j)
	if err != nil {
		log.Println(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, BytesCombine([]byte("1"), filename[:], file))
	if err != nil {
		log.Println(err)
		return
	} else {
		bar.Add(1)
	}
}

// Del File
func fld(j string, bar *progressbar.ProgressBar) {
	var filename [256]byte
	for ii, jj := range j {
		filename[ii] = byte(jj)
	}

	err := c.WriteMessage(websocket.BinaryMessage, BytesCombine([]byte("0"), filename[:]))
	if err != nil {
		log.Println(err)
		return
	} else {
		bar.Add(1)
	}
}

func main() {
	defer c.Close()

	filelist, err := GetAllFile(config.Path)
	if err != nil {
		fmt.Println("Err:", err)
		return
	}

	delfilelist := Arrcmp(filelist, cache)
	filelistcmp := Arrcmp(cache, filelist)

	bar1 := progressbar.Default(int64(len(delfilelist)))
	bar2 := progressbar.Default(int64(len(filelistcmp)))

	index := 0
	for {
		if index < len(delfilelist) {
			fld(delfilelist[index], bar1)

			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if string(msg) == "next" {
				index++
			} else {
				break
			}
		} else {
			break
		}
	}

	index = 0
	for {
		if index < len(filelistcmp) {
			flc(filelistcmp[index], bar2)

			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if string(msg) == "next" {
				index++
			} else {
				break
			}
		} else {
			break
		}
	}

	file, err := os.OpenFile("./cache.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("File Open Error", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	jsonStr, _ := json.Marshal(filelist)
	write.WriteString(string(jsonStr))
	write.Flush()
}
