package cralwer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
	CrawlerTime string                   `json:"crawler_time"`
}

type Crawler struct {
	crawlerName string
}

// CrawlerWeiBo 爬取微博热榜信息
func (c Crawler) CrawlerWeiBo() (Result, error) {
	var content []map[string]interface{}
	timeout := 10 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	mUrl := "https://m.weibo.cn/api/container/getIndex?containerid=106003type%3D25%26t%3D3%26disable_hot%3D1%26filter_type%3Drealtimehot&title=%E5%BE%AE%E5%8D%9A%E7%83%AD%E6%90%9C&extparam=seat%3D1%26pos%3D0_0%26dgr%3D0%26mi_cid%3D100103%26cate%3D10103%26filter_type%3Drealtimehot%26c_type%3D30%26display_time%3D1638445376%26pre_seqid%3D52252862&luicode=10000011&lfid=231583"
	req, err := http.NewRequest("GET", mUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	if err != nil {
		fmt.Println("CrawlerWeiBo http.NewRequest err:", err)
		return Result{HotName: "新浪微博"}, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("client.Do err:", err)
		return Result{HotName: "新浪微博"}, err
	}
	if res.StatusCode != 200 {
		log.Printf(" CrawlerWeiBo status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "新浪微博"}, err
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
		return Result{HotName: "新浪微博"}, err
	}
	cardGroup := j.Get("data").Get("cards").GetIndex(0).Get("card_group").MustArray()
	for _, val := range cardGroup {
		title := val.(map[string]interface{})["desc"]
		href := fmt.Sprintf("https://s.weibo.com/weibo?q=%%23%s%%23", title)
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}
	result := Result{"新浪微博", content, time.Now().Format("2006-01-02 15:04:05")}

	return result, nil
}

// CrawlerZhiHu 爬取知乎热榜信息
func (c Crawler) CrawlerZhiHu() (Result, error) {
	var content []map[string]interface{}
	url := "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true"
	timeout := 5 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerZhiHu http.NewRequest err:", err)
		return Result{HotName: "知乎热榜"}, err
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
		return Result{HotName: "知乎热榜"}, err
	}
	body, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println("")
		return Result{HotName: "知乎热榜"}, err
	}
	dataJson := j.Get("data")
	dataArr := j.Get("data").MustArray()
	for index := range dataArr {
		info := dataJson.GetIndex(index)
		title := info.Get("target").Get("title_area").Get("text").MustString()
		href := info.Get("target").Get("link").Get("url").MustString()
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}

	return Result{"知乎热榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

// CrawlerTieBa 爬取贴吧热榜
func (c Crawler) CrawlerTieBa() (Result, error) {
	var content []map[string]interface{}
	url := "https://tieba.baidu.com/hottopic/browse/topicList"
	timeout := time.Second * 10
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerTieBa http.NewRequest err:", err)
		return Result{HotName: "贴吧"}, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerTieBa client.Do err:", err)
		return Result{HotName: "贴吧"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)

	str, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("CrawlerTieBa simplejson.NewJson err:", err)
		return Result{HotName: "贴吧"}, err
	}
	topicList := j.Get("data").Get("bang_topic").Get("topic_list")
	topicArr := topicList.MustArray()
	for index := range topicArr {
		title := topicList.GetIndex(index).Get("topic_name").MustString()
		href := topicList.GetIndex(index).Get("topic_url").MustString()
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}
	return Result{"贴吧", content, time.Now().Format("2006-01-02 15:04:05")}, nil

}

// CrawlerDouBan 爬取豆瓣热榜
func (c Crawler) CrawlerDouBan() (Result, error) {
	var content []map[string]interface{}
	url := "https://www.douban.com/group/explore"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerDouBan http.NewRequest err:", err)
		return Result{HotName: "豆瓣热榜"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Host", "www.douban.com")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerDouBan client.Do err:", err)
		return Result{HotName: "豆瓣热榜"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("CrawlerDouBan status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "豆瓣热榜"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerDouBan goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "豆瓣热榜"}, err
	}
	doc.Find(".channel-item").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h3 a").Text()
		href, boolHref := s.Find("h3 a").Attr("href")
		if boolHref {
			content = append(content, map[string]interface{}{"title": title, "href": href})
		}

	})
	return Result{"豆瓣热榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

// CrawlerTianYa 爬取天涯热榜
func (c Crawler) CrawlerTianYa() (Result, error) {
	var content []map[string]interface{}
	url := "http://bbs.tianya.cn/hotArticle.jsp"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerTianYa http.NewRequest err:", err)
		return Result{HotName: "天涯热榜"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Host", "bbs.tianya.cn")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerTianYa client.Do err:", err)
		return Result{HotName: "天涯热榜"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("CrawlerTianYa status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "天涯热榜"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerTianYa goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "天涯热榜"}, err
	}
	doc.Find(".mt5 table tbody tr").Slice(1, -1).Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("td[class=td-title] a").Text()
		href := "http://bbs.tianya.cn" + selection.Find("td[class=td-title] a").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": title, "href": href})
	})

	return Result{"天涯热榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil

}

// CrawlerGithub 爬取github trending
func (c Crawler) CrawlerGithub() (Result, error) {
	var content []map[string]interface{}
	url := "https://github.com/trending"
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerDouBan http.NewRequest err:", err)
		return Result{HotName: "GitHub Trending"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("Referer", "https://github.com/explore")
	req.Header.Add("Host", "github.com")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerDouBan client.Do err:", err)
		return Result{HotName: "GitHub Trending"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("CrawlerDouBan status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "GitHub Trending"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerDouBan goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "GitHub Trending"}, err
	}
	doc.Find("article[class=Box-row]").Each(func(i int, selection *goquery.Selection) {
		title := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(selection.Find("h2 a").Text()), "\n", ""), " ", "")
		href := "https://github.com/" + selection.Find("h2 a").AttrOr("href", "")
		describe := strings.TrimSpace(selection.Find("p").Text())
		content = append(content, map[string]interface{}{"title": title + "<---->" + describe, "href": href})

	})
	return Result{"GitHub Trending", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

// CrawlerWangYiYun 获取网易云音乐
func (c Crawler) CrawlerWangYiYun() (Result, error) {
	var content []map[string]interface{}
	url := "https://music.163.com/discover/toplist?id=19723756"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerWangYiYun http.NewRequest err:", err)
		return Result{HotName: "云音乐飙升榜"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	req.Header.Add("authority", "music.163.com")
	req.Header.Add("Referer", "https://music.163.com/")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerWangYiYun client.Do err:", err)
		return Result{HotName: "云音乐飙升榜"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("CrawlerWangYiYun status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "云音乐飙升榜"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerWangYiYun goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "云音乐飙升榜"}, err
	}

	doc.Find("div[id=song-list-pre-cache] ul[class=f-hide] li").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("a").Text()
		href := "https://music.163.com/#%s" + selection.Find("a").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": title, "href": href})

	})
	return Result{"云音乐飙升榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

func (c Crawler) CrawlerCSDN() (Result, error) {
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

	return Result{"CSDN热榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

//CrawlerWeread 获取微信读书热榜
func (c Crawler) CrawlerWeread() (Result, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	var content []map[string]interface{}

	url := "https://weread.qq.com/web/category/rising"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerWeread http.NewRequest err:", err)
		return Result{HotName: "微信读书飙升榜"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("CrawlerWangYiYun client.Do err:", err)
		return Result{HotName: "微信读书飙升榜"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("CrawlerWangYiYun status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "微信读书飙升榜"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("CrawlerWangYiYun goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "微信读书飙升榜"}, err
	}
	doc.Find(".ranking_content_bookList li[class=wr_bookList_item]").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("p[class=wr_bookList_item_title]").Text()
		href := "https://weread.qq.com" + selection.Find("a[class=wr_bookList_item_link]").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": title, "href": href})
	})

	return Result{"微信读书飙升榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

// Crawler52PoJie 吾爱破解
func (c Crawler) Crawler52PoJie() (Result, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	var content []map[string]interface{}

	url := "https://www.52pojie.cn/forum.php?mod=guide&view=hot"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Crawler52PoJie http.NewRequest err:", err)
		return Result{HotName: "吾爱破解"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Crawler52PoJie client.Do err:", err)
		return Result{HotName: "吾爱破解"}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err:", err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("Crawler52PoJie status code error: %d %s", res.StatusCode, res.Status)
		return Result{HotName: "吾爱破解"}, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Crawler52PoJie goquery.NewDocumentFromReader err:", err)
		return Result{HotName: "吾爱破解"}, err
	}
	doc.Find("#threadlist .bm_c tbody").Each(func(i int, selection *goquery.Selection) {
		gbkTitle := selection.Find("tr th a[class=xst]").Text()
		title, err := io.ReadAll(transform.NewReader(bytes.NewReader([]byte(gbkTitle)), simplifiedchinese.GBK.NewDecoder()))
		if err != nil {
			fmt.Println("52 gbk to utf8 err:", err)
		}
		href := "https://www.52pojie.cn/forum.php?mod=guide&view=hot" + selection.Find("tr th a").AttrOr("href", "")
		content = append(content, map[string]interface{}{"title": string(title), "href": href})
	})

	return Result{"吾爱破解", content, time.Now().Format("2006-01-02 15:04:05")}, nil
}

// CrawlerDouYin 抖音
func (c Crawler) CrawlerDouYin() (Result, error) {
	var content []map[string]interface{}
	url := "https://www.douyin.com/aweme/v1/web/hot/search/list/?device_platform=webapp&aid=6383&channel=channel_pc_w" +
		"eb&detail_list=1&source=6&pc_client_type=1&version_code=170400&version_name=17.4.0&cookie_enabled=true&screen" +
		"_width=1440&screen_height=900&browser_language=en&browser_platform=MacIntel&browser_name=Chrome&browser_" +
		"version=107.0.0.0&browser_online=true&engine_name=Blink&engine_version=107.0.0.0&os_name=Mac+OS&os_version=" +
		"10.15.7&cpu_core_num=8&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=100&webid" +
		"=7168107943232308770&msToken=x872gxuQF3TKQoShjH0dOxcP5vMWOtp9vE3gAhYMVfvklclynZ5uOj8KsIw_WML0fzol" +
		"EFqOw4NUSbVwMCIEqGNEs0tFx7hyogm9SI43HP4f__VTIc-mZgOCAjMj7A==&X-Bogus=DFSzswVOS1UANtnuS8c4f37TlqCw"
	timeout := 5 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CrawlerDouYin http.NewRequest err:", err)
		return Result{HotName: "抖音热榜"}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	req.Header.Add("authority", "www.douyin.com")
	req.Header.Add("referer", "https://www.douyin.com/hot")
	req.Header.Add("sec-ch-ua-platform", "macOS")
	req.Header.Add("accept", "application/json, text/plain, */*")
	res, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("err", err)
		}
	}(res.Body)
	if err != nil {
		fmt.Println("CrawlerDouYin client.Do err:", err)
		return Result{HotName: "抖音热榜"}, err
	}
	body, _ := io.ReadAll(res.Body)
	j, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println("")
		return Result{HotName: "抖音热榜"}, err
	}
	dataJson := j.Get("data").Get("word_list")
	dataArr := j.Get("data").Get("word_list").MustArray()
	for index := range dataArr {
		info := dataJson.GetIndex(index)
		title := info.Get("word").MustString()
		href := "https://www.douyin.com/hot/" + info.Get("sentence_id").MustString()
		content = append(content, map[string]interface{}{"title": title, "href": href})
	}

	return Result{"抖音热榜", content, time.Now().Format("2006-01-02 15:04:05")}, nil
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
	fmt.Println("开始时间：", time.Now().Format("2006-01-02 15:04:05"))
	allCrawler := []string{"CrawlerWeiBo", "CrawlerZhiHu", "CrawlerTieBa", "CrawlerDouBan", "CrawlerTianYa",
		"CrawlerGithub", "CrawlerWangYiYun", "CrawlerCSDN", "CrawlerWeread", "Crawler52PoJie", "CrawlerDouYin"}
	cr := make(chan Result, len(allCrawler))
	for _, value := range allCrawler {
		wg.Add(1)
		fmt.Println("开始抓取" + value)
		crawler := Crawler{value}
		go ExecGetData(crawler, cr)
	}
	wg.Wait()
	fmt.Print("抓取结束：", time.Now().Format("2006-01-02 15:04:05"))
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
	ticker := time.NewTicker(60 * 10 * time.Second)

	for range ticker.C {
		RunCrawlerAndWrite()
	}

}
