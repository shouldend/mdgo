package mdgo

import (
	"net/http"
	"os"
	"testing"
)

var (
	c, _ = NewHtmlConverter(
		StartId("layout-content"),
		TagIgnore("div", "page-tools"),
		DefaultLang("php"),
	)
)

func TestHtml(t *testing.T) {
	resp, err := http.Get("http://php.net/manual/en/function.file-get-contents.php")
	if err != nil {
		t.Log(err)
		return
	}
	defer resp.Body.Close()
	if err = c.ConvertIO(resp.Body, os.Stdout); err != nil {
		t.Log(err)
	}
}
