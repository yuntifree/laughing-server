package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

//GetMD5Hash return hex md5 of text
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

//GenUUID gen uuid
func GenUUID() string {
	u := uuid.NewV4()
	return u.String()
}

//GenSaltPasswd calc new password with salt
func GenSaltPasswd(password, salt string) string {
	return GetMD5Hash(password + salt)
}

//GenSalt gen 32 byte hex string
func GenSalt() string {
	uuid := GenUUID()
	return strings.Join(strings.Split(uuid, "-"), "")
}

//Rand gen random
func Rand() int32 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Int31()
}

//Randn gen random between 0 and n
func Randn(n int32) int32 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Int31n(n)
}

//Sha1 gen hex sha1 of content
func Sha1(content string) string {
	hash := sha1.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

//HmacSha1 gen hex sha1 of conten with key
func HmacSha1(content, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(content))
	return hex.EncodeToString(mac.Sum(nil))
}
