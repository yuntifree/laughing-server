package spider

import "testing"

func Test_IsInstagramDst(t *testing.T) {
	m := map[string]bool{
		"https://instagram.com/p/BSupJ2YlQwP/":       true,
		"https://www.instagram.com/p/BR5Hfm5Dv81/":   true,
		"https://www.instagram.com/p/BR5Hfm5Dv81/xx": false,
	}
	for k, v := range m {
		if IsInstagramDst(k) != v {
			t.Errorf("IsInstagramDst check failed:%s", k)
		}
	}
}

func Test_IsFacebookDst(t *testing.T) {
	m := map[string]bool{
		"https://m.facebook.com/story.php?story_fbid=787516061405226&id=149180321905473": true,
		"https://www.facebook.com/gioliofficialpage/videos/733025296854303/":             true,
		"https://m.facebook.com/story.php?story_fbid":                                    false,
	}
	for k, v := range m {
		if IsFacebookDst(k) != v {
			t.Errorf("IsFacebookDst check failed:%s", k)
		}
	}
}

func Test_IsMusicallyDst(t *testing.T) {
	m := map[string]bool{
		"https://www.musical.ly/v/MzcyNTEyMTYyNTc3MzA1OTkyMjMyOTY.html":   true,
		"https://www.musical.ly/v/MzcyNTEyMTYyNTc3MzA1OTkyMjMyOTY.html?x": false,
	}
	for k, v := range m {
		if IsMusicallyDst(k) != v {
			t.Errorf("IsMusicallyDst check failed:%s", k)
		}
	}
}

func Test_IsYoutubeDst(t *testing.T) {
	m := map[string]bool{
		"https://youtu.be/6Nxc-3WpMbg":                               true,
		"https://www.youtube.com/watch?v=6Nxc-3WpMbg":                true,
		"https://m.youtube.com/watch?feature=youtu.be&v=6Nxc-3WpMbg": true,
		"https://m.youtube.com/watch?feature=youtu&v=6Nxc-3WpMbg":    false,
	}
	for k, v := range m {
		if IsYoutubeDst(k) != v {
			t.Errorf("IsYoutubeDst check failed:%s", k)
		}
	}
}
