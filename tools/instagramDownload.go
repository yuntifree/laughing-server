package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

var urls = []string{
	//"https://instagram.com/p/BSupJ2YlQwP/",
	"https://www.instagram.com/p/BR5Hfm5Dv81/",
}

type VideoInfo struct {
	bid,
	caption,
	thumbUrl,
	videoUrl string
	height,
	width,
	duration int
}

func checkerr(e error) bool {
	if e != nil {
		panic(e)
		return true
	}
	return false
}

func getVideoInfo(url string) *VideoInfo {
	vinfo := &VideoInfo{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if checkerr(err) {
		return vinfo
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if checkerr(err) {
		return vinfo
	}

	content := string(bytes)

	pos1 := strings.Index(content, "window._sharedData = {")
	if pos1 == -1 {
		log.Printf("Can't find window._shareData, CHECK!!\n")
		return vinfo
	}

	str := content[pos1+21:]
	pos2 := strings.Index(str, ";</script>")

	jsonObj, err := simplejson.NewJson([]byte(str[:pos2]))
	if checkerr(err) {
		return vinfo
	}

	postPage := jsonObj.Get("entry_data").Get("PostPage")
	if postPage == nil {
		log.Printf("Parse json error 1, CHECK!!\n")
		return vinfo
	}

	media := postPage.GetIndex(0).Get("graphql").Get("shortcode_media")
	if media == nil {
		log.Printf("Parse json error 2, CHECK!!\n")
		return vinfo
	}

	isVideo, _ := media.Get("is_video").Bool()
	if !isVideo {
		log.Printf("It's not video\n")
		return vinfo
	}

	vinfo.thumbUrl, _ = media.Get("display_url").String()
	vinfo.videoUrl, _ = media.Get("video_url").String()
	vinfo.width, _ = media.Get("dimensions").Get("width").Int()
	vinfo.height, _ = media.Get("dimensions").Get("height").Int()
	vinfo.bid, _ = media.Get("shortcode").String()

	return vinfo
}

func processUrl(url string) {
	vinfo := getVideoInfo(url)
	log.Printf("video info: %v", vinfo)
}

func main() {
	for _, v := range urls {
		processUrl(v)
	}
}
