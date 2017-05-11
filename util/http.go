package util

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

//HTTPRequest return response body of http request
func HTTPRequest(url, reqbody string) (string, error) {
	return HTTPRequestWithHeaders(url, reqbody, map[string]string{})
}

//HTTPRequestWithHeaders return response body of http request with headers
func HTTPRequestWithHeaders(url, reqbody string, headers map[string]string) (string, error) {
	client := &http.Client{Timeout: time.Second * 5}
	method := "GET"
	if len(reqbody) > 0 {
		method = "POST"
	}
	req, err := http.NewRequest(method, url, strings.NewReader(reqbody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rspbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(rspbody), nil
}

//DownFile download file from url
func DownFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("downFile get failed:%s %v", url, err)
		return []byte(""), err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}
