package facebook

import (
	"fmt"
	"laughing-server/util"
	"log"

	simplejson "github.com/bitly/go-simplejson"
)

const (
	baseurl   = "https://graph.facebook.com/v2.9/"
	imgurl    = "https://graph.facebook.com/"
	imgsuffix = "/picture?type=large"
)

//GetName get name from fb
func GetName(id, token string) (name string, err error) {
	url := fmt.Sprintf("%s%s?access_token=%s", baseurl, id, token)
	log.Printf("url:%s", url)
	resp, err := util.HTTPRequest(url, "")
	if err != nil {
		return
	}

	log.Printf("resp:%s", resp)
	js, err := simplejson.NewJson([]byte(resp))
	if err != nil {
		return
	}

	name, err = js.Get("name").String()
	return
}

//GenHeadurl generate fb head url
func GenHeadurl(id string) string {
	return fmt.Sprintf("%s/%s/%s", imgurl, id, imgsuffix)
}
