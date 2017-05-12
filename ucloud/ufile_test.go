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
