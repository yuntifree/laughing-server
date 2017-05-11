package ucloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	//Bucket  = "laugh"
	//host    = "laugh.us-ca.ufileos.com"
	Bucket    = "chatcat"
	host      = "http://chatcat.hk.ufileos.com"
	cdn       = "http://chatcat.ufile.ucloud.com.cn"
	pubkey    = "ZeZGjUnEz+A7gxeVGxTNUhwDLGJj21SPTqOmSvPN+0WtGwvhDMmseg=="
	privkey   = "cdb4fe689528a582425fa96e235e094e9da75f3f"
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
	client := &http.Client{Timeout: time.Second * 5}
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
