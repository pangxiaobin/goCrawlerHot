package main

import (
	"fmt"
	"goCrawlerHot/cralwer"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

func hotDataHandler(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Body.Close err:", err)
		}
	}(r.Body)
	data := r.URL.Query()
	fmt.Println(data.Get("name"))
	fmt.Println(data.Get("age"))
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	//answer := `{"status": "ok"}`
	fileName := filepath.Join(baseDir, "result.json")
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Read file err, err =", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("file cloe err:", err)
		}
	}(file)
	jsonData, err := io.ReadAll(file)
	_, err = w.Write(jsonData)
	if err != nil {
		fmt.Println("w.Write err:", err)
	}
}

var baseDir string

func init() {
	baseDir, _ := os.Getwd()
	fmt.Println(baseDir)
}

func main() {
	go cralwer.RunTicker()
	//单独写回调函数
	http.HandleFunc("/hot-data", hotDataHandler)

	http.Handle("/layui/", http.StripPrefix("/layui/", http.FileServer(http.Dir("./html/layui/"))))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		temFilePath := filepath.Join(baseDir, "html", "index.html")
		temp, err := template.ParseFiles(temFilePath)
		if err != nil {
			fmt.Println("create template failed err:", err)
			return
		}
		// 利用给定数据渲染模板，并将结果写入w
		err = temp.Execute(writer, "")
		if err != nil {
			fmt.Println("temp.Execute err:", err)
		}
	})

	// addr：监听的地址
	// handler：回调函数
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("http.ListenAndServe err:", err)
	}
}
