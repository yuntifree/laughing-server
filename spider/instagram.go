package spider

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

//GetInstagram get video info from instagram
func GetInstagram(url string) *VideoInfo {
	vinfo := &VideoInfo{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	checkerr(err)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)
	checkerr(err)
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	checkerr(err)

	content := string(bytes)

	pos1 := strings.Index(content, "window._sharedData = {")
	if pos1 == -1 {
		log.Printf("Can't find window._shareData, CHECK!!\n")
		return vinfo
	}

	str := content[pos1+21:]
	pos2 := strings.Index(str, ";</script>")

	jsonObj, err := simplejson.NewJson([]byte(str[:pos2]))
	checkerr(err)

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
