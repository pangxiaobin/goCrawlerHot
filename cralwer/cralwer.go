package cralwer

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Result struct {
	HotName     string                   `json:"hot_name"`
	Content     []map[string]interface{} `json:"content"`
	CrawlerTime time.Time                `json:"crawler_time"`
}

type Crawler struct {
	crawlerName string
}

// CrawlerWeiBo 爬取微博热榜信息
func (c Crawler) CrawlerWeiBo() Result {
	timeout := 10 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	mUrl := "https://m.weibo.cn/api/container/getIndex?containerid=106003type%3D25%26t%3D3%26disable_hot%3D1%26filter_type%3Drealtimehot&title=%E5%BE%AE%E5%8D%9A%E7%83%AD%E6%90%9C&extparam=seat%3D1%26pos%3D0_0%26dgr%3D0%26mi_cid%3D100103%26cate%3D10103%26filter_type%3Drealtimehot%26c_type%3D30%26display_time%3D1638445376%26pre_seqid%3D52252862&luicode=10000011&lfid=231583"
	req, err := http.NewRequest("GET", mUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	if err != nil {
		fmt.Println("CrawlerWeiBo http.NewRequest err:", err)
		return Result{}
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("client.Do err:", err)
	}
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("CrawlerWeiBo close err:", err)
		}
	}(res.Body)
	str, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println(" simplejson.NewJson err:", err)
	}
	var content []map[string]interface{}
	cardGroup := j.Get("data").Get("cards").GetIndex(0).Get("card_group").MustArray()
	for _, val := range cardGroup {
		title := val.(map[string]interface{})["desc"]
		href := fmt.Sprintf("https://s.weibo.com/weibo?q=%%23%s%%23", title)
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}
	result := Result{"新浪微博", content, time.Now()}

	return result
}

// CrawlerZhiHu 爬取知乎热榜信息
func (c Crawler) CrawlerZhiHu() Result {
	url := "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true"
	timeout := 5 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerZhiHu http.NewRequest err:", err)
		return Result{}
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("path", "/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true")
	req.Header.Add("x-api-version", "3.0.76")
	req.Header.Add("x-requested-with", "fetch")
	res, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err", err)
		}
	}(res.Body)
	if err != nil {
		fmt.Println("CrawlerZhiHu client.Do err:", err)
		return Result{}
	}
	var content []map[string]interface{}
	body, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println("")
		return Result{}
	}
	dataJson := j.Get("data")
	dataArr := j.Get("data").MustArray()
	for index := range dataArr {
		info := dataJson.GetIndex(index)
		title := info.Get("target").Get("title_area").Get("text").MustString()
		href := info.Get("target").Get("link").Get("url").MustString()
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}

	return Result{"知乎热榜", content, time.Now()}
}

// CrawlerTieBa 爬取贴吧热榜
func (c Crawler) CrawlerTieBa() Result {
	url := "https://tieba.baidu.com/hottopic/browse/topicList"
	timeout := time.Second * 10
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerTieBa http.NewRequest err:", err)
		return Result{}
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerTieBa client.Do err:", err)
		return Result{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	var content []map[string]interface{}
	str, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("CrawlerTieBa simplejson.NewJson err:", err)
		return Result{}
	}
	topicList := j.Get("data").Get("bang_topic").Get("topic_list")
	topicArr := topicList.MustArray()
	for index := range topicArr {
		title := topicList.GetIndex(index).Get("topic_name").MustString()
		href := topicList.GetIndex(index).Get("topic_url").MustString()
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}
	return Result{"贴吧", content, time.Now()}

}

// CrawlerDouBan 爬取豆瓣热榜
func (c Crawler) CrawlerDouBan() Result {
	url := "https://www.douban.com/group/explore"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerDouBan http.NewRequest err:", err)
		return Result{}
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Host", "www.douban.com")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerDouBan client.Do err:", err)
		return Result{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerDouBan goquery.NewDocumentFromReader err:", err)
		return Result{}
	}
	var content []map[string]interface{}
	doc.Find(".channel-item").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h3 a").Text()
		href, boolHref := s.Find("h3 a").Attr("href")
		if boolHref {
			content = append(content, map[string]interface{}{"title": title, "href": href})
		}

	})
	return Result{"豆瓣热榜", content, time.Now()}
}

// CrawlerTianYa 爬取天涯热榜
func (c Crawler) CrawlerTianYa() Result {
	url := "http://bbs.tianya.cn/hotArticle.jsp"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerTianYa http.NewRequest err:", err)
		return Result{}
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Host", "bbs.tianya.cn")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerTianYa client.Do err:", err)
		return Result{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerTianYa goquery.NewDocumentFromReader err:", err)
		return Result{}
	}
	var content []map[string]interface{}
	doc.Find(".mt5 table tbody tr").Slice(1, -1).Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("td[class=td-title] a").Text()
		href := "http://bbs.tianya.cn" + selection.Find("td[class=td-title] a").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": title, "href": href})
	})

	return Result{"天涯热榜", content, time.Now()}

}

// CrawlerGithub 爬取github trending
func (c Crawler) CrawlerGithub() Result {
	url := "https://github.com/trending"
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerDouBan http.NewRequest err:", err)
		return Result{}
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Referer", "https://github.com/explore")
	req.Header.Add("Host", "github.com")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerDouBan client.Do err:", err)
		return Result{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerDouBan goquery.NewDocumentFromReader err:", err)
		return Result{}
	}
	var content []map[string]interface{}
	doc.Find("article[class=Box-row]").Each(func(i int, selection *goquery.Selection) {
		title := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(selection.Find("h1 a").Text()), "\n", ""), " ", "")
		href := "https://github.com/" + selection.Find("h1 a").AttrOr("href", "")
		describe := strings.TrimSpace(selection.Find("p").Text())
		content = append(content, map[string]interface{}{"title": title + "<---->" + describe, "href": href})

	})
	return Result{"GitHub Trending", content, time.Now()}
}

// CrawlerWangYiYun 获取网易云音乐
func (c Crawler) CrawlerWangYiYun() Result {
	url := "https://music.163.com/discover/toplist?id=19723756"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerWangYiYun http.NewRequest err:", err)
		return Result{}
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("authority", "music.163.com")
	req.Header.Add("Referer", "https://music.163.com/")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerWangYiYun client.Do err:", err)
		return Result{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerWangYiYun goquery.NewDocumentFromReader err:", err)
		return Result{}
	}
	var content []map[string]interface{}

	doc.Find("div[id=song-list-pre-cache] ul[class=f-hide] li").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("a").Text()
		href := "https://music.163.com/#%s" + selection.Find("a").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": title, "href": href})

	})
	return Result{"云音乐飙升榜", content, time.Now()}
}

func (c Crawler) CrawlerCSDN() Result {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	pages := [4]string{"0", "1", "2", "3"}
	var content []map[string]interface{}
	for _, page := range pages {
		url := "https://blog.csdn.net/phoenix/web/blog/hot-rank?page=" + page + "&pageSize=25&type="
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("CrawlerCSDN http.NewRequest err:", err)
			continue
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("CrawlerCSDN client.Do err:", err)
			continue
		}
		str, _ := io.ReadAll(res.Body)
		j, err := simplejson.NewJson(str)
		if err != nil {
			fmt.Println("CrawlerCSDN simplejson.NewJson err:", err)
			continue
		}
		func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("err:", err)
			}
		}(res.Body)
		dataJson := j.Get("data")
		dataArr := j.Get("data").MustArray()
		for index := range dataArr {
			info := dataJson.GetIndex(index)
			title := info.Get("articleTitle").MustString()
			href := info.Get("articleDetailUrl").MustString()
			content = append(content, map[string]interface{}{"title": title, "href": href})
		}
	}

	return Result{"CSDN热榜", content, time.Now()}
}

func ExecGetData(c Crawler, cr chan Result) {
	reflectValue := reflect.ValueOf(c)
	crawler := reflectValue.MethodByName(c.crawlerName)
	data := crawler.Call(nil)
	originData := data[0].Interface().(Result)
	cr <- originData
	wg.Done()
}

var wg sync.WaitGroup

//RunCrawlerAndWrite  爬取数据并写入文件
func RunCrawlerAndWrite() {
	// 文件创建
	fmt.Println("开始时间：", time.Now())
	allCrawler := []string{"CrawlerWeiBo", "CrawlerZhiHu", "CrawlerTieBa", "CrawlerDouBan", "CrawlerTianYa",
		"CrawlerGithub", "CrawlerWangYiYun", "CrawlerCSDN"}
	cr := make(chan Result, len(allCrawler))
	for _, value := range allCrawler {
		wg.Add(1)
		fmt.Println("开始抓取" + value)
		crawler := Crawler{value}
		go ExecGetData(crawler, cr)
	}
	wg.Wait()
	fmt.Print("抓取结束：", time.Now())
	close(cr)
	var resultInfo []Result
	for val := range cr {
		resultInfo = append(resultInfo, val)
	}
	baseDir, _ := os.Getwd()
	fileLock := new(sync.RWMutex)
	resultPath := filepath.Join(baseDir, "result.json")
	fileLock.Lock()
	file, _ := os.OpenFile(resultPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("err", err)
		}
	}(file)
	output, _ := json.Marshal(&resultInfo)
	file.Write(output)
	fileLock.Unlock()
}

func RunTicker() {
	RunCrawlerAndWrite()
	// 定时任务, 2分钟爬取一次
	ticker := time.NewTicker(60 * 2 * time.Second)

	for range ticker.C {
		RunCrawlerAndWrite()
	}

}
