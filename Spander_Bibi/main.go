package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

/*爬取B站电影评论
先摸到主评论源代码，一页二十个
发现每个主评论都有一个自己的“root”，搭配一个网址可以进入子评论源代码页面
开始乱爬
爬取时间很长，爬一页花了大概十几分钟...也有第一页个个子评论几百条的原因？我进的子评论源代码页面一页仅十个子评论，可以访问太花时间了
	我自个为了不被B站墙写了个很长的time.sleep也是原因之一  悲    也许我的方式也是问题之一？
*/

func HttpGet(url string) (result string, err error) {
	/*req_url := "http://httpbin.org/get"
	proxyAddr := "http://125.46.0.62:53281"
	proxy, _ := url2.Parse(proxyAddr)
	transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
	client := &http.Client{Transport: transport}
	原先想写一个代理地址，然后跑起来报错了，就暂时放着了，最后改了下其他地方，正常跑了，就把这给忘了
	*/
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		os.Exit(12)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.42")
	req.Header.Add("Cookie", "buvid3=388BADD4-5CD4-8B72-7D1B-4837C063323054403infoc; i-wanna-go-back=-1; _uuid=B9BF8F8C-F3FA-174C-774D-C155A3398C3B60141infoc; buvid4=88F25495-4E77-B635-E54E-3C835EB777ED65693-022081122-m4T90WVXeaiq1MO3QR4GHg==; buvid_fp_plain=undefined; nostalgia_conf=-1; hit-dyn-v2=1; blackside_state=1; is-2022-channel=1; b_nut=100; CURRENT_BLACKGAP=0; LIVE_BUVID=AUTO6316640153364175; hit-new-style-dyn=0; CURRENT_FNVAL=4048; rpdid=|(u)luk)~)ml0J'uYY)Y|)RYu; CURRENT_QUALITY=0; fingerprint=15e4a723c2594a16fc4c0ebdb092b284; sid=7t6tflk3; bsource=search_bing; buvid_fp=f6462f1ecebb9aaa743aedc555331694; bp_video_offset_32885952=736820262778961900; PVID=3; b_lsid=C1A83316_184EBC4647D; innersign=0; b_ut=7")
	req.Header.Add("Referer", "https://www.bilibili.com/bangumi/play/ss12548?theme=movie&spm_id_from=333.337.0.0")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		os.Exit(resp.StatusCode)
	}
	if err != nil {
		log.Println(err)
		os.Exit(15)
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
		result += string(buf[:n])
	}
	time.Sleep(5 * time.Second)
	return
}

//func HttpGet(url string) (result string, err error) {
//	//fmt.Println("====================")
//	res, err := http.Get(url)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer res.Body.Close()
//	if res.StatusCode != 200 {
//		fmt.Println("Http Status code:", res.StatusCode)
//		os.Exit(res.StatusCode)
//		return
//	}
//	buf := make([]byte, 4096)
//	for {
//		n, err2 := res.Body.Read(buf)
//		if n == 0 {
//			break
//		}
//		if err2 != nil && err2 != io.EOF {
//			err = err2
//			return
//		}
//		result += string(buf[:n])
//	}
//	return
//} 啊B的防护性

func Work(url string, q int) bool {
	res, _ := HttpGet(url) //  得到起始主评论网页源代码
	regParentId := regexp.MustCompile(`"root_str":"(?s:(.*?))","parent_str"`)
	ParentId := regParentId.FindAllStringSubmatch(res, -1) //获得主评论的“root”
	if len(ParentId) == 0 {
		//页数超出实际页数时，捕捉不到root，返回空数组
		return true
	}
	Pid := RemoveDuplicates(ParentId) //删去重复获得的root
	//fmt.Println(Pid)  大主页20个主评论root getdaze！
	//fmt.Println(ParentId)
	f, err := os.Create("第 " + strconv.Itoa(q) + "页评论" + ".html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var result string
	for i := 1; i < len(Pid); i++ {
		fmt.Println("========================" + strconv.Itoa(i)) //测试用
		for j := 1; ; j++ {
			fmt.Println("===============" + strconv.Itoa(j)) //同上
			//url1 := "https://api.bilibili.com/x/v2/reply/reply?=jQuery17207579796859218155_1670386496905&jsonp=jsonp&pn=" + strconv.Itoa(j) + "&type=1&oid=21071819&ps=10&root=" + Pid[i]
			url1 := "https://api.bilibili.com/x/v2/reply/reply?=jQuery17207579796859218155_1670386496905&jsonp=jsonp&pn=" + strconv.Itoa(j) + "&type=1&oid=21071819&ps=10&root=" + Pid[i]
			//构建初始子评论源代码页面
			res1, _ := HttpGet(url1)
			//fmt.Println(res1)
			fmt.Println("正常运行中。。。预计运行几百次吧") //测试
			kid := kidConnent(res1)
			if len(kid) == 1 {
				//页数“j”如果超出了实际页数，会到达一个只有主评论源代码的页面...所以len=1
				fmt.Println("到停止啦！！！！！！") //测试
				break
			}
			for _, s := range kid {
				result += s
				result += "\n"
			}
		}
	}
	f.WriteString(result)
	return false
}

func kidConnent(res string) []string {
	reg := regexp.MustCompile(`"content":{"message":"(?s:(.*?)),"plat":0,`)
	regKidConnent := reg.FindAllStringSubmatch(res, -1)
	kid := make([]string, len(regKidConnent))
	for i, s := range regKidConnent {
		kid[i] = s[1]
	}
	return kid
}

func RemoveDuplicates(slice1 [][]string) []string {
	slice := make([]string, len(slice1))
	for i, strings := range slice1 {
		slice[i] = strings[1]
	}
	sort.Strings(slice)
	i := 0
	var j int
	for {
		if j > len(slice)-1 {
			break
		}
		for j = i + 1; j < len(slice) && slice[i] == slice[j]; j++ {
		}
		slice = append(slice[:i+1], slice[j:]...)
		i++
	}

	return slice
} //将获得的二维数组转为一维数组，并进行去重

func main() {
	for i := 1; ; i++ {
		url := "https://api.bilibili.com/x/v2/reply/main?=jQuery17207579796859218155_1670386496895&jsonp=jsonp&next=" + strconv.Itoa(i) + "&type=1&oid=21071819&mode=3&plat=1"
		flag := Work(url, i)
		if flag {
			break
		}
	}
}
