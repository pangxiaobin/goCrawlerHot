package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

type Result struct {
	HotName     string                   `json:"hot_name"`
	Content     []map[string]interface{} `json:"content"`
	CrawlerTime time.Time                `json:"crawler_time"`
}

var baseDir string

func init() {
	baseDir, _ := os.Getwd()
	fmt.Println(baseDir)
}

func main() {
	//go cralwer.RunTicker()
	http.Handle("/layui/", http.StripPrefix("/layui/", http.FileServer(http.Dir("./html/layui/"))))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		temFilePath := filepath.Join(baseDir, "html", "index.html")
		htmlByte, err := ioutil.ReadFile(temFilePath)
		if err != nil {
			fmt.Println("read html failed, err:", err)
			return
		}
		// 自定义一个夸人的模板函数
		addNum := func(arg int) (int, error) {
			return arg + 1, nil
		}
		// 采用链式操作在Parse之前调用Funcs添加自定义的kua函数
		tmpl, err := template.New("hello").Funcs(template.FuncMap{"addNum": addNum}).Parse(string(htmlByte))
		if err != nil {
			fmt.Println("create template failed, err:", err)
			return
		}

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
		str, _ := io.ReadAll(file)
		var hotData []Result
		err = json.Unmarshal(str, &hotData)
		if err != nil {
			fmt.Println("json Unmarshal err:", err)
		}
		// 利用给定数据渲染模板，并将结果写入w
		err = tmpl.Execute(writer, hotData)
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
