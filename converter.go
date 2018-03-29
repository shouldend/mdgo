package mdgo

import "io"

type Converter interface {
	Init(...OptionFunc) error
	Convert(string) (string, error)
	ConvertIO(io.Reader, io.Writer) error
	base() *baseConverter
}

type baseConverter struct {
	Converter
	holdComment         bool
	tagReplaceMap       map[string]map[string]string
	tagIgnoreMap        map[string]map[string]bool
	startTag, startAttr string
	startId             string
	lang                string
	langMap             map[string]string
}

func newBase() *baseConverter {
	return &baseConverter{
		tagReplaceMap: make(map[string]map[string]string),
		tagIgnoreMap:  make(map[string]map[string]bool),
		langMap:       make(map[string]string),
	}
}

func (c *baseConverter) base() *baseConverter {
	return (*baseConverter)(c)
}
