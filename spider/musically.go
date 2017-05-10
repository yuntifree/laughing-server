package spider

import (
	"net/http"

	simplejson "github.com/bitly/go-simplejson"
)

//GetMusically get video info from musically
func GetMusically(url string) *VideoInfo {
	info := &VideoInfo{}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	checkerr(err)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)

	checkerr(err)

	defer resp.Body.Close()

	jsonObj, err := simplejson.NewFromReader(resp.Body)

	checkerr(err)

	succ, err := jsonObj.Get("success").Bool()

	checkerr(err)

	if !succ {
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
