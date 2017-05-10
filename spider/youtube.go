package spider

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

type kvPair map[string]string

//GetYoutube get info from youtube
func GetYoutube(key string) *VideoInfo {
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
	keymap := kvPair{}
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
	pairList := []kvPair{}
	for _, v := range formatList {
		//itag=22&url=xxx&sig=xxx
		pairArr := strings.Split(v, "&")
		pair := kvPair{}
		for _, v1 := range pairArr {
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
