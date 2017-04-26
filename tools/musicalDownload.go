package main

import (
	"log"
	"net/http"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

const (
	infoBase = "https://www.musical.ly/rest/v2/musicals/shareInfo?key="
)

type VideoInfo struct {
	bid,
	caption,
	thumbUrl,
	videoUrl string
	height,
	width,
	duration int
}

var urls = []string{
	"https://www.musical.ly/v/MzcyNTEyMTYyNTc3MzA1OTkyMjMyOTY.html",
}

func checkerr(e error) bool {
	if e != nil {
		panic(e)
		return true
	}
	return false
}

func getVideoInfo(url string) *VideoInfo {
	info := &VideoInfo{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if checkerr(err) {
		return info
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	jsonObj, err := simplejson.NewFromReader(resp.Body)

	if checkerr(err) {
		return info
	}

	succ, err := jsonObj.Get("success").Bool()

	if checkerr(err) || !succ {
		return info
	}

	result := jsonObj.Get("result")
	info.bid, _ = result.Get("bid").String()
	info.caption, _ = result.Get("caption").String()
	info.duration, _ = result.Get("duration").Int()
	info.height, _ = result.Get("height").Int()
	info.width, _ = result.Get("width").Int()
	info.thumbUrl, _ = result.Get("thumbnailUri").String()
	info.videoUrl, _ = result.Get("videoUri").String()

	return info
}

func processUrl(url string) {
	pos1 := strings.LastIndex(url, "/")
	pos2 := strings.LastIndex(url, ".html")

	infoUrl := infoBase + string(url[pos1+1:pos2])

	vinfo := getVideoInfo(infoUrl)

	log.Printf("video info: %v", vinfo)
}

func main() {
	for _, url := range urls {
		processUrl(url)
	}
}
