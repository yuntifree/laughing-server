package ucloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"laughing-server/util"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	//Bucket ucloud file bucket
	//Bucket = "laugh"
	//host   = "http://laugh.us-ca.ufileos.com"
	Bucket    = "chatcat"
	host      = "http://chatcat.hk.ufileos.com"
	cdn       = "http://laugh.us-ca.ufileos.com"
	pubkey    = "qVEFK9wRsdWqMols6VCfijDQ/dYp+xK4BHUChSj4Aauwg2QcsI6tyQ=="
	privkey   = "ef547cd0481874c18258e460f9d6a1582bd1d57e"
	thumbnail = "?iopcmd=thumbnail&type=4&width=400"
)

func genSign(content, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(content))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

//GetCdnURL get cdn down url
func GetCdnURL(filename string) string {
	return cdn + "/" + filename
}

//GetThumbnailURL get cdn thumbnail url
func GetThumbnailURL(filename string) string {
	return cdn + "/" + filename + thumbnail
}

//PutFile put file to bucket
func PutFile(bucket, filename string, buf []byte) bool {
	str := "PUT" + "\n\n\n\n/" + bucket + "/" + filename
	sign := genSign(str, privkey)
	method := "PUT"
	client := &http.Client{Timeout: time.Second * 30}
	url := host + "/" + filename
	req, err := http.NewRequest(method, url, bytes.NewReader(buf))
	length := len(buf)
	auth := "UCloud " + pubkey + ":" + sign
	req.Header.Set("Content-Length", strconv.Itoa(length))
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("PutFile do http request failed:%v", err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Printf("PutFile failed code:%d", resp.StatusCode)
		return false
	}
	return true
}

//GenUploadToken generate upload token
func GenUploadToken(format string) (url string, auth string) {
	filename := util.GenUUID() + format
	url = host + "/" + filename
	str := "PUT" + "\n\n\n\n/" + Bucket + "/" + filename
	sign := genSign(str, privkey)
	auth = "UCloud " + pubkey + ":" + sign
	return
}
