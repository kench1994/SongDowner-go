package main

import (
	"fmt"
	"net/http"
	"net/url"
	"compress/gzip"
	"io"
	"io/ioutil"
	"encoding/json"
	"os"
	"sync"
)

const (
	SuggestionUrl = "http://sug.music.baidu.com/info/suggestion"
	Fmlink        = "http://music.baidu.com/data/music/fmlink"
)

func main() {
	//	fmt.Println("Hello World!")
	fmt.Println("Program start.")
	//query := url.Values{}

	musicName := "晴天"
	query := SetMusicListBody(string(musicName))
	res,err := DownloadString(SuggestionUrl,query)
	if err != nil {
		fmt.Println("获取音乐列表时出错：",err)
		return
	}
	var dat map[string]interface{}
	err = json.Unmarshal([]byte(res), &dat)
	if err != nil {
		fmt.Println("反序列化JSON时出错:",err)
		return
	}
	if _,ok := dat["data"]; ok == false{
		fmt.Println("没有找到音乐资源:",string(musicName))
		return
	}
	songid := dat["data"].(map[string]interface{})["song"].([]interface{})[0].(map[string]interface{})["songid"].(string)



	query = SetMusicLinkBody(songid,"flac")
	res ,err = DownloadString(Fmlink,query)
	if err != nil{
		fmt.Println("获取音乐文件时出错：",err)
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(res,&data)
	if code,ok:= data["errorCode"]; (ok && code.(float64) == 22005) || err != nil {
		fmt.Println("解析音乐文件时出错：",err)
		return
	}
	songlink := data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songLink"].(string)
	fmt.Println("songlink: ",songlink)
	r := []rune(songlink)
	if len(r) < 10 {
		fmt.Println("没有无损音乐地址:",string(musicName))
		return
	}
	songname :=  data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songName"].(string)
	artistName :=  data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["artistName"].(string)
	//fmt.Println("songName: ",songname,"artistName: ",artistName)


	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	var path string
	if  os.IsPathSeparator('\\') {
		path = "\\"
	}else{
		path = "/"
	}
	dir, _ := os.Getwd()
	dir = dir +path+"songs"
	if _,err := os.Stat(dir);err != nil{
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			fmt.Println("创建目录失败：",err)
			return
		}
	}
	filename := dir + path + songname+" - "+artistName+".flac"
	fmt.Println(filename)
	//go func() {
	fmt.Println("正在下载 ", songname," ......")
	defer waitGroup.Done()

	songRes ,err:= http.Get(songlink)
	if err != nil {
		fmt.Println("下载文件时出错：",songlink)
		return
	}
	songFile,err := os.Create(filename)
	written,err := io.Copy(songFile,songRes.Body)
	if err != nil {
		fmt.Println("保存音乐文件时出错：",err)
		return
	}
	fmt.Println(songname,"下载完成,文件大小：",fmt.Sprintf("%.2f", float64(written)/(1024*1024)),"MB")
	waitGroup.Wait()

	//}()

	//waitGroup.Wait()

}

func SetMusicLinkBody(songId string,songType string)(queryVaules url.Values){
	queryVaules = url.Values{}
	queryVaules.Set("songIds",songId)
	queryVaules.Set("type",songType)
	return
}

func SetMusicListBody(songName string)(queryVaules url.Values){
	queryVaules = url.Values{}
	queryVaules.Set("word", string(songName))
	queryVaules.Set("version","2")
	queryVaules.Set("from","0")
	return
}

func DownloadString(remoteUrl string, queryVaules url.Values) (body []byte, err error) {
	client := &http.Client{}
	body = nil
	uri, err := url.Parse(remoteUrl)
	if err != nil {
		return
	}
	if queryVaules != nil {
		values := uri.Query()
		for k, v := range values {
			queryVaules[k] = v
		}
		uri.RawQuery = queryVaules.Encode()
	}
	fmt.Println(uri)

	reqest, err := http.NewRequest("GET", uri.String(), nil)
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Encoding", "gzip, deflate")
	reqest.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Host", uri.Host)
	reqest.Header.Add("Referer", uri.String())
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	response, err := client.Do(reqest)
	defer response.Body.Close()
	if err != nil {
		fmt.Println("response error")
		return
	}

	if response.StatusCode == 200 {
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ := gzip.NewReader(response.Body)
			for {
				buf := make([]byte, 1024)
				n, err := reader.Read(buf)

				if err != nil && err != io.EOF {
					panic(err)
				}

				if n == 0 {
					break
				}
				body = append(body, buf...)
			}
		default:
			body, _ = ioutil.ReadAll(response.Body)

		}
	}
	return
}