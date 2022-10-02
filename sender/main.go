package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/schollz/progressbar/v3"
)

var config struct {
	Server string `json:"server"`
	Token  string `json:"token"`
	Path   string `json:"path"`
}

var cache []string

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
}

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

func upload(path, filename string) bool {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("files", filename)

	file, _ := os.Open(path)
	defer file.Close()

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(fmt.Sprintf("%s/backup?token=%s&name=%s", config.Server, url.QueryEscape(config.Token), url.QueryEscape(filename)), contentType, bodyBuffer)
	if err != nil {
		fmt.Println("Err", err)
	}
	defer resp.Body.Close()

	resp_body, _ := io.ReadAll(resp.Body)

	switch string(resp_body) {
	case "ok":
		return true
	case "token":
		fmt.Println("\nErr:", "Token Error")
		os.Exit(1)
	default:
		fmt.Println("\nErr:", string(resp_body))
	}
	return false
}

func GetAllFile(pathname string) ([]string, error) {
	result := []string{}

	fis, err := os.ReadDir(pathname)
	if err != nil {
		fmt.Println("ReadErr:", err)
		return result, err
	}

	for _, fi := range fis {
		fullname := path.Join(pathname, fi.Name())
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

func main() {
	filelist, err := GetAllFile(config.Path)
	if err != nil {
		fmt.Println("Err:", err)
		return
	}

	filelistcmp := Arrcmp(cache, filelist)

	bar := progressbar.Default(int64(len(filelistcmp)))

	var flist []string
	for _, j := range filelistcmp {
		for i := 1; i < 3; i++ {
			if upload(config.Path+j, j) {
				bar.Add(1)
				break
			} else {
				if i == 2 {
					flist = append(flist, j)
				}
			}
		}
	}

	fail := len(flist)
	if fail != 0 {
		fmt.Println("Err:", flist)
	}
	fmt.Printf("total: %d, success: %d, fail: %d", len(filelist), len(filelistcmp)-fail, fail)

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
