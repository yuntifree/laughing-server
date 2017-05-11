package ucloud

import (
	"io/ioutil"
	"testing"
)

func Test_PutFile(t *testing.T) {
	filename := "ufile_test.go"
	bucket := "chatcat"
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("read file failed:%s %v", filename, err)
	}
	if !PutFile(bucket, filename, buf) {
		t.Errorf("PutFile failed")
	}
}
