package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	simplejson "github.com/bitly/go-simplejson"
)

var urls = []string{
	"https://m.facebook.com/story.php?story_fbid=787516061405226&id=149180321905473",
	//"https://www.facebook.com/gioliofficialpage/videos/733025296854303/",
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

func checkerr(e error) {
	if e != nil {
		panic(e)
	}
}

func getVideoInfo(url string) *VideoInfo {
	vinfo := &VideoInfo{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	checkerr(err)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)
	checkerr(err)
	defer resp.Body.Close()

	d, err := goquery.NewDocumentFromReader(resp.Body)
	checkerr(err)

	div := d.Find(".widePic")
	video := div.Find("div").First()
	jsonText, _ := video.Attr("data-store")
	style, _ := video.Find("i").First().Attr("style")

	jsonObj, err := simplejson.NewJson([]byte(jsonText))
	checkerr(err)

	vType, _ := jsonObj.Get("type").String()
	if vType != "video" {
		log.Printf("This is not a video\n")
		return vinfo
	}

	vinfo.bid, _ = jsonObj.Get("videoID").String()
	vinfo.videoUrl, _ = jsonObj.Get("src").String()
	vinfo.width, _ = jsonObj.Get("width").Int()
	vinfo.height, _ = jsonObj.Get("height").Int()

	// find thumb
	pos1 := strings.Index(style, "(")
	if pos1 == -1 {
		log.Printf("The video has no thumb\n")
	} else {
		pos2 := strings.Index(style, ")")

		vinfo.thumbUrl = style[pos1+2 : pos2-1]
	}

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
