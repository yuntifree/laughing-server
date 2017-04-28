package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	imgBase  = "https://i.ytimg.com/vi/%s/hqdefault.jpg"
	infoBase = "http://youtube.com/get_video_info?video_id="
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

type Pair map[string]string

var urls = []string{
	"https://youtu.be/6Nxc-3WpMbg",
	//"https://www.youtube.com/watch?v=6Nxc-3WpMbg",
	//"https://m.youtube.com/watch?feature=youtu.be&v=6Nxc-3WpMbg",
}

func checkerr(e error) {
	if e != nil {
		panic(e)
	}
}

func getVideoInfo(key string) *VideoInfo {
	vinfo := &VideoInfo{}
	vinfo.thumbUrl = fmt.Sprintf(imgBase, key)

	client := &http.Client{}
	req, err := http.NewRequest("GET", infoBase+key, nil)

	checkerr(err)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36")

	resp, err := client.Do(req)

	checkerr(err)

	defer resp.Body.Close()

	// get raw video info from youtube api
	raw, err := ioutil.ReadAll(resp.Body)
	checkerr(err)

	// split into keys
	keylist := strings.Split(string(raw), "&")
	keymap := Pair{}
	for _, v := range keylist {
		pair := strings.Split(v, "=")
		val, _ := url.QueryUnescape(pair[1])
		keymap[pair[0]] = val
	}

	// we need: video_id,url_encoded_fmt_stream_map,adaptive_fmts,dashmpd
	v, ok := keymap["url_encoded_fmt_stream_map"]
	if !ok {
		log.Println("url_encoded_fmt_stream_map Not Found")
		return vinfo
	}

	formatList := strings.Split(v, ",")
	pairList := []Pair{}
	for _, v := range formatList {
		//itag=22&url=xxx&sig=xxx
		pair_str := strings.Split(v, "&")
		pair := Pair{}
		for _, v1 := range pair_str {
			p := strings.Split(v1, "=")
			val, _ := url.QueryUnescape(p[1])
			pair[p[0]] = val
		}
		pairList = append(pairList, pair)
	}

	for _, v := range pairList {
		// 18 for mp4, 17 for 3gp
		if itag, ok := v["itag"]; ok && itag == "18" {
			if videoUrl, ok := v["url"]; ok {
				vinfo.videoUrl = videoUrl
			}
		}
	}

	return vinfo
}

func processUrl(v string) {

	u, err := url.Parse(v)
	checkerr(err)

	key := string(u.Path[1:])
	if key == "watch" {
		q := u.Query()
		key = q.Get("v")
	}

	vinfo := getVideoInfo(key)
	log.Printf("video info: %v\n", vinfo)
}

func main() {
	for _, url := range urls {
		processUrl(url)
	}
}
