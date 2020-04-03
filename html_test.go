package mdgo

import (
	"net/http"
	"os"
	"testing"
)

var (
	c, _ = NewHtmlConverter(
		StartTag("div", "blog-content-box"),
		TagIgnore("div", "article-info-box"),
		TagIgnore("div", "more-toolbox"),
		TagIgnore("div", "person-messagebox"),
		DefaultLang("java"),
		MapLang("language-xml", "xml"),
		MapLang("language-java", "java"),
	)
)

func TestHtml(t *testing.T) {
	resp, err := http.Get("https://blog.csdn.net/yiifaa/article/details/72860927")
	if err != nil {
		t.Log(err)
		return
	}
	defer resp.Body.Close()
	file, _ := os.OpenFile("ApplicationContext.md", os.O_RDWR|os.O_CREATE, 0644)
	if err = c.ConvertIO(resp.Body, file); err != nil {
		t.Log(err)
	}
}
