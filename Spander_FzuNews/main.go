package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)
//福大要闻读取

func GetNumber(clickid string, owner string) string {//观看人数部分，通过获得变化的“clickid”和“owner”进行访问，获得响应
	var result string
	//fmt.Println("https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=" + clickid + "&owner=" + owner + "&clicktype=wbnews")
	resp, _ := http.Get("https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=" + clickid + "&owner=" + owner + "&clicktype=wbnews")
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, _ := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		result += string(buf[:n])
	}
	fmt.Println(result)
	return result
}


func GetNext(url string) string {
	reg := regexp.MustCompile(`<h1 class="highlight next"><a href="(?s:(.*?)).htm"><span>`)
	next := reg.FindAllStringSubmatch(url, -1) //.htm部分不能直接爬貌似，后改成只爬数字部分，如26529.htm的26529
	return next[0][1]
} //返回下一页的网址改变部分-->MakeUrl

func MakeUrl(url string) string {
	return "https://news.fzu.edu.cn/info/1011/" + url + ".htm"
} //和GetNext合作创建下一页的网址-->Work

func HttpGet(url string) (result string, err error) {
	resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			fmt.Println("=============================读取网页完成=====================")
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
		result += string(buf[:n])
	}
	return
} //获取网页内容

func Work(url string, i int) (string, bool) { //i是主函数循环对应的i，为了便于创建文件名
	resp, err := HttpGet(url)
	if err != nil {
		fmt.Println("HttpGet err:", err)
	}
	//开始爬特定内容
	f, err := os.Create("第 " + strconv.Itoa(i) + " 篇" + ".html")
	if err != nil {
		fmt.Println("Create err:", err)
	}
	//标题
	regTitle := regexp.MustCompile(`<div class="nav01">(?s:(.*?))</h3>`)
	title := regTitle.FindAllStringSubmatch(resp, -1)
	//发布日期
	regDay := regexp.MustCompile(`<span>发布日期:(?s:(.*?))</span>`)
	Day := regDay.FindAllStringSubmatch(resp, -1)
	flag := false
	if Day[0][1] == "  2022-09-01" {
		flag = true
	}
	//作者
	regAuthor := regexp.MustCompile(`<span>作者：(?s:(.*?))</span>`)
	Author := regAuthor.FindAllStringSubmatch(resp, -1)
	//观看人数
	regClicked := regexp.MustCompile(`<script>_showDynClicks\("wbnews", [0-9]{10},`)
	clicked := regClicked.FindAllStringSubmatch(resp, -1)
	clicked[0][0] = clicked[0][0][33:43]
	fmt.Println(clicked[0][0])
	regOwner := regexp.MustCompile(`, [0-9]{5}\)</script></span>`)
	owner := regOwner.FindAllStringSubmatch(resp, -1)
	owner[0][0] = owner[0][0][2:7]
	fmt.Println(owner[0][0])
	Read := "阅读人数：" + GetNumber(owner[0][0], clicked[0][0])
	//正文
	regText := regexp.MustCompile(`<p class="vsbcontent_start"><strong>融媒中心讯/</strong>(?s:(.*?))</div></div><div id="div_vote_id"></div>`)
	text := regText.FindAllStringSubmatch(resp, -1)
	var result string
	for _, strings := range title {
		result += strings[1]
	}
	result += "\n"
	for _, strings := range Day {
		result += strings[1]
	}
	result += "\n"
	for _, strings := range Author {
		result += strings[1]
	}
	result += "\n"
	fmt.Println("++++++++++++" + Read)
	result += Read
	result += "\n"
	for _, strings := range text {
		result += strings[1]
	}
	result += "\n"
	f.WriteString(result)
	f.Close()

	return GetNext(resp), flag
}

func main() {
	startUrl := "https://news.fzu.edu.cn/info/1011/" + "26553.htm"
	url := startUrl
	for i := 1; ; i++ {
		urlPart, stop := Work(url, i)
		url = MakeUrl(urlPart)
		if stop {
			break
		}
	}
}
