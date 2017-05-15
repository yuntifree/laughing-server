package ucloud

import (
	"io/ioutil"
	"testing"
)

func Test_PutFile(t *testing.T) {
	filename := "ufile.go"
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("read file failed:%s %v", filename, err)
	}
	if !PutFile(Bucket, filename, buf) {
		t.Errorf("PutFile failed")
	}
}

func Test_GenHeadurl(t *testing.T) {
	m := map[string]string{
		"": "",
		"http://img.yunxingzh.com/e5fc5ff5-302b-43fc-b2ef-26b5652a6479.jpg": "http://img.yunxingzh.com/e5fc5ff5-302b-43fc-b2ef-26b5652a6479.jpg",
		"e5fc5ff5-302b-43fc-b2ef-26b5652a6479.jpg":                          cdn + "/e5fc5ff5-302b-43fc-b2ef-26b5652a6479.jpg",
	}
	for k, v := range m {
		if GenHeadurl(k) != v {
			t.Errorf("GenHeadurl failed:%s-%s:%s", k, v, GenHeadurl(k))
		}
	}
}
