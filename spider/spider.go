package spider

import "regexp"

const (
	instagram = `^https\:\/\/(www\.){0,1}instagram\.com\/p\/(\w)+\/$`
	facebook  = `^https\:\/\/m\.facebook\.com\/story\.php\?story_fbid\=(\d)+\&id\=(\d)+$|^https\:\/\/www\.facebook\.com\/gioliofficialpage\/videos\/(\d)+/$`
	youtube   = `^https\:\/\/youtu\.be\/(\w|\-)+$|^https\:\/\/www\.youtube\.com\/watch\?v\=(\w|\-)+$|^https\:\/\/m\.youtube\.com\/watch\?feature\=youtu\.be\&v\=(\w|\-)+$`
	musically = `^https\:\/\/www\.musical\.ly\/v\/(\w)+\.html$`
)

//VideoInfo video info
type VideoInfo struct {
	bid,
	caption,
	thumbUrl,
	videoUrl string
	height,
	width,
	duration int
}

func checkerr(e error) {
	if e != nil {
		panic(e)
	}
}

func checkPattern(dst, pattern string) bool {
	matched, err := regexp.MatchString(pattern, dst)
	if err == nil && matched {
		return true
	}
	return false
}

//IsInstagramDst check Instagram video url
func IsInstagramDst(dst string) bool {
	return checkPattern(dst, instagram)
}

//IsFacebookDst check Facebook video url
func IsFacebookDst(dst string) bool {
	return checkPattern(dst, facebook)
}

//IsMusicallyDst check Musically video url
func IsMusicallyDst(dst string) bool {
	return checkPattern(dst, musically)
}

//IsYoutubeDst check Youtube video url
func IsYoutubeDst(dst string) bool {
	return checkPattern(dst, youtube)
}
