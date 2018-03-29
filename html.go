package mdgo

import (
	"fmt"
	"io"

	"strings"

	"bytes"

	"regexp"

	"golang.org/x/net/html"
)

// htmlConverter
// convert html to md
type htmlConverter struct {
	*baseConverter
	// options
	quoteLevel int
	isRaw      bool
}

// NewHtmlConverter - create an converter
func NewHtmlConverter(funcs ...OptionFunc) (Converter, error) {
	converter := &htmlConverter{
		baseConverter: newBase(),
	}
	for _, f := range funcs {
		if err := f(converter); err != nil {
			return nil, err
		}
	}
	return converter, nil
}

// next - next operation
func (c *htmlConverter) next(w io.Writer, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			c.text(w, child)
		case html.CommentNode:
			c.comment(w, child)
		case html.ElementNode:
			var (
				tag       = strings.ToLower(child.Data)
				needParse = true
			)
			if val, ok := c.tagIgnoreMap[tag]; ok {
				for _, attr := range child.Attr {
					if _, exists := val[attr.Val]; exists {
						needParse = false
						break
					}
				}
				if _, exists := val["_all"]; exists {
					needParse = false
				}
			}
			if needParse {
				c.element(w, tag, child)
			}
			continue
		default:
			//c.next(w, child)
		}
	}
}

// untagged functions
// element - process element for all tags
func (c *htmlConverter) element(w io.Writer, tag string, node *html.Node) {
	switch tag {
	case "a":
		c.a(w, node)
	case "b", "strong":
		c.b(w, node)
	case "br":
		c.br(w, node)
	case "code":
		c.code(w, node)
	case "del":
		c.del(w, node)
	case "em", "i":
		c.em(w, node)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		c.h(w, node)
	case "head":
		return
	case "hr":
		c.hr(w, node)
	case "image":
		c.image(w, node)
	case "li":
		c.li(w, node)
	case "nav":
		return
	case "p":
		c.p(w, node)
	case "pre":
		c.pre(w, node)
	case "quote":
		c.quote(w, node)
	case "span", "div":
		c.blockElement(tag, w, node)
	case "table":
		c.table(w, node)
	case "ul":
		c.ul(w, node)
	default:
		c.next(w, node)
	}
}

// text - text block
func (c *htmlConverter) text(w io.Writer, node *html.Node) {
	if c.isRaw {
		fmt.Fprint(w, node.Data)
		return
	}
	// trim spaces
	_ = c
	if strings.TrimSpace(node.Data) == "" {
		return
	}
	//s := strings.Trim(node.Data, "\t\r\n")
	s := regexp.MustCompile(`[[:space:]][[:space:]]*`).ReplaceAllString(strings.Trim(node.Data, "\t\r\n"), " ")
	s = replacer.Replace(s)
	// block quote
	s = strings.NewReplacer("\n", "\n"+strings.Repeat("> ", c.quoteLevel)).Replace(s)
	fmt.Fprint(w, s)
}

// comment - "<!-- sth here -->"
func (c *htmlConverter) comment(w io.Writer, node *html.Node) {
	if c.holdComment {
		fmt.Fprintf(w, "<!--%s-->\n", node.Data)
	}
}

// a - link, "<a href=..."
func (c *htmlConverter) a(w io.Writer, node *html.Node) {
	if c.isRaw {
		c.next(w, node)
		return
	}
	c.inc(w, node, "[", "](%s)", getAttr(node, "href"))
}

// b - "b" or "strong"
func (c *htmlConverter) b(w io.Writer, node *html.Node) {
	c.inc(w, node, "**", "**")
}

// br - means a new line
func (c *htmlConverter) br(w io.Writer, node *html.Node) {
	_ = node
	_ = c
	fmt.Fprint(w, "\n\n")
}

// code - always "code", "tt"
func (c *htmlConverter) code(w io.Writer, node *html.Node) {
	if getParentKey(node) != "pre" {
		if node.FirstChild == node.LastChild && node.FirstChild != nil {
			// standard type for code, always an inline block element
			switch node.FirstChild.Type {
			case html.TextNode:
				temp := c.isRaw
				c.isRaw = true
				fmt.Fprintf(w, "`%s`", node.FirstChild.Data)
				c.isRaw = temp
			case html.ElementNode:
				if strings.ToLower(node.Data) == "a" {
					// link
					c.inc(w, node.FirstChild, "[`", "`](%s)", getAttr(node, "href"))
					return
				} else {
					c.pre(w, node)
				}
			default:
				return
			}
			return
		}
	}
	// uncommon type, may be just use as a pre
	c.pre(w, node)
}

// del - strickout
func (c *htmlConverter) del(w io.Writer, node *html.Node) {
	c.inc(w, node, "~~", "~~")
}

// em - "em" or "i"
func (c *htmlConverter) em(w io.Writer, node *html.Node) {
	c.inc(w, node, "_", "_")
}

// h - header context
func (c *htmlConverter) h(w io.Writer, node *html.Node) {
	count := node.Data[1] - '0'
	c.inc(w, node, fmt.Sprintf("\n%s ", strings.Repeat("#", int(count))), "\n")
}

// hr - separator, always "--"
func (c *htmlConverter) hr(w io.Writer, node *html.Node) {
	_ = node
	fmt.Fprint(w, "\n\n--\n\n")
}

// img - image link
func (c *htmlConverter) image(w io.Writer, node *html.Node) {
	fmt.Fprintf(w, "![%s](%s)\n", getAttr(node, "alt"), getAttr(node, "src"))
}

// li - list
func (c *htmlConverter) li(w io.Writer, node *html.Node) {
	c.inc(w, node, "* ", "\n")
}

// p - always for a new line
func (c *htmlConverter) p(w io.Writer, node *html.Node) {
	line := "\n" + strings.Repeat("> ", c.quoteLevel)
	c.inc(w, node, strings.Repeat(line, 2), "\n\n")
}

// pre - a code block
func (c *htmlConverter) pre(w io.Writer, node *html.Node) {
	c.isRaw = true
	originLang := c.lang
	for _, attr := range node.Attr {
		if lang, exists := c.langMap[attr.Val]; exists {
			c.lang = lang
		}
	}
	defer func() {
		c.isRaw = false
		c.lang = originLang
	}()
	// if pre is an wrapper for code, next
	if node.FirstChild == node.LastChild &&
		node.FirstChild != nil &&
		strings.ToLower(node.FirstChild.Data) == "code" {
		c.pre(w, node.FirstChild)
		return
	}
	c.inc(w, node, fmt.Sprintf("\n\n```%s\n", c.lang), "\n```\n")
}

// quote - block quote, actually, quote will be processed in text
func (c *htmlConverter) quote(w io.Writer, node *html.Node) {
	c.quoteLevel++
	c.inc(w, node, "\n", "\n")
	c.quoteLevel--
}

// blockElement - process block element like div, span, etc.
func (c *htmlConverter) blockElement(tag string, w io.Writer, node *html.Node) {
	if spanRep, ok := c.tagReplaceMap[tag]; ok {
		for _, attr := range node.Attr {
			if tag, exists := spanRep[attr.Key]; exists {
				c.element(w, tag, node)
				return
			}
		}
		if tag, exists := spanRep["_all"]; exists {
			c.element(w, tag, node)
			return
		}
	}
	c.next(w, node)
}

// style
func (c *htmlConverter) style(w io.Writer, node *html.Node) {

}

// script
func (c *htmlConverter) script(w io.Writer, node *html.Node) {
	return
}

// table - process table things
func (c *htmlConverter) table(w io.Writer, node *html.Node) {
	// multi-line property for one table should be considered
	if node.FirstChild == nil {
		return
	}
	var split = " ---- "
	switch getAttr(node, "align") {
	case "left":
		split = " :---- "
	case "right":
		split = " ----: "
	case "center":
		split = " :----: "
	}
	// thead, tbody and tfoot
	var (
		thead, tfoot []string
		tbody        [][]string
		maxLen       = 0
	)
	getTr := func(node *html.Node) (ret []string, tp int) {
		defer func() {
			if maxLen < len(ret) {
				maxLen = len(ret)
			}
		}()
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			switch strings.ToLower(child.Data) {
			case "th":
				tp = 1
				fallthrough
			case "td":
				// as tag may be exists, just decode
				data := new(bytes.Buffer)
				for sc := child.FirstChild; sc != nil; sc = sc.NextSibling {
					onelineFormat(data, sc)
				}
				ret = append(ret, data.String())
			}
		}
		return
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch strings.ToLower(child.Data) {
		case "thead":
			if child.FirstChild != nil {
				thead, _ = getTr(child.FirstChild)
			}
		case "tfoot":
			if child.FirstChild != nil {
				tfoot, _ = getTr(child.FirstChild)
			}
		case "tbody":
			for sc := child.FirstChild; sc != nil; sc = sc.NextSibling {
				if sr, tb := getTr(sc); sr != nil {
					if tb == 1 {
						thead = sr
					} else {
						tbody = append(tbody, sr)
					}
				}
			}
		}
	}
	// nil head
	if thead == nil {
		for i := 0; i < maxLen; i++ {
			thead = append(thead, " ")
		}
	}
	// head here
	fmt.Fprintf(w, "\n\n| %s |\n", strings.Join(thead, " | "))
	// split here
	for i := 0; i < maxLen; i++ {
		fmt.Fprintf(w, "| %s ", split)
	}
	fmt.Fprint(w, "|\n")
	// should print body here
	for _, body := range tbody {
		fmt.Fprintf(w, "| %s |\n", strings.Join(body, " | "))
	}
	// foot here
	if tfoot != nil {
		fmt.Fprintf(w, "| %s |\n", strings.Join(tfoot, " | "))
	}
}

// ul
func (c *htmlConverter) ul(w io.Writer, node *html.Node) {
	c.inc(w, node, "\n\n", "\n")
}

// inc - print for a tag is interrupted by a next operation
func (c *htmlConverter) inc(w io.Writer, node *html.Node, beg, end string, v ...interface{}) {
	fmt.Fprint(w, beg)
	c.next(w, node)
	fmt.Fprintf(w, end, v...)
}

// tools
// getAttr - get attr for an element node of html
func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return strings.ToLower(attr.Val)
		}
	}
	return ""
}

// getParentKey - get key for parent node
func getParentKey(node *html.Node) string {
	switch node.Parent.Type {
	case html.ElementNode:
		return strings.ToLower(node.Parent.Data)
	default:
		return ""
	}
}

// nodeSearchByTag - search exact node by tag
func nodeSearchByTag(node *html.Node, tag, attr string) *html.Node {
	if node == nil {
		return nil
	}
	if strings.ToLower(node.Data) == tag {
		if attr == "_all" {
			return node
		}
		if getAttr(node, "class") == attr {
			return node
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n := nodeSearchByTag(child, tag, attr); n != nil {
			return n
		}
	}
	return nil
}

// nodeSearchById - search exact node by id
func nodeSearchById(node *html.Node, id string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == id {
				return node
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n := nodeSearchById(child, id); n != nil {
			return n
		}
	}
	return nil
}

// oneLineFormat - format node into one line
func onelineFormat(w io.Writer, node *html.Node) {
	if node == nil {
		return
	}
	switch node.Type {
	case html.TextNode:
		fmt.Fprint(w, strings.Replace(replacer.Replace(node.Data), "\n", "<br>", -1))
		return
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		if tag == "br" || tag == "hr" {
			fmt.Fprintf(w, "<%s \\>", tag)
		} else {
			fmt.Fprintf(w, "<%s>", tag)
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				onelineFormat(w, child)
			}
			fmt.Fprintf(w, "</%s>", tag)
		}
	}
}

// printNodeChild
func printNodeChild(node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		fmt.Printf("%+v\n", child)
	}
}

// Exports
func (c *htmlConverter) Convert(content string) (string, error) {
	node, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if c.startId != "" {
		node = nodeSearchById(node, c.startId)
		if node == nil {
			return "", fmt.Errorf("can not find element for id(%s)", c.startId)
		}
	} else if c.startTag != "" {
		node = nodeSearchByTag(node, c.startTag, c.startAttr)
		if node == nil {
			return "", fmt.Errorf("can not find element for tag(%s-%s)", c.startTag, c.startAttr)
		}
	}
	c.next(buf, node)
	return buf.String(), nil
}

// Exports
func (c *htmlConverter) ConvertIO(r io.Reader, w io.Writer) error {
	node, err := html.Parse(r)
	if err != nil {
		return err
	}
	if c.startId != "" {
		node = nodeSearchById(node, c.startId)
		if node == nil {
			return fmt.Errorf("can not find element for id(%s)", c.startId)
		}
	} else if c.startTag != "" {
		node = nodeSearchByTag(node, c.startTag, c.startAttr)
		if node == nil {
			return fmt.Errorf("can not find element for tag(%s-%s)", c.startTag, c.startAttr)
		}
	}
	c.next(w, node)
	return nil
}
