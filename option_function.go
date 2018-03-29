package mdgo

import "errors"

type OptionFunc func(Converter) error

// HoldComment - hold comments for document
func HoldComment(b bool) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		c.base().holdComment = b
		return nil
	}
}

// TagReplace - a tag may be use for another tag
func TagReplace(tag, attr string, rep string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		if tag == "" || rep == "" {
			return errors.New("nil tag not allowed")
		}
		if attr == "" {
			attr = "_all"
		}
		if _, ok := c.base().tagReplaceMap[tag]; ok {
			c.base().tagReplaceMap[tag][attr] = rep
		} else {
			c.base().tagReplaceMap[tag] = map[string]string{attr: rep}
		}
		return nil
	}
}

// StartTag - tag to start parse at
func StartTag(tag, attr string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		if tag == "" {
			return errors.New("nil tag not allowed")
		}
		if attr == "" {
			attr = "_all"
		}
		c.base().startTag = tag
		c.base().startAttr = attr
		return nil
	}
}

// StartId - unlike tag, id is unique
func StartId(id string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		if id == "" {
			return errors.New("nil id not allowed")
		}
		c.base().startId = id
		return nil
	}
}

// TagIgnore - which tag to be ignore while parsing
func TagIgnore(tag, attr string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		if tag == "" {
			return errors.New("nil tag not allowed")
		}
		if attr == "" {
			attr = "_all"
		}
		if _, ok := c.base().tagIgnoreMap[tag]; ok {
			c.base().tagIgnoreMap[tag][attr] = true
		} else {
			c.base().tagIgnoreMap[tag] = map[string]bool{attr: true}
		}
		return nil
	}
}

// DefaultLang - Set default programming language for coding block
func DefaultLang(lang string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		c.base().lang = lang
		return nil
	}
}

// MapLang - Set language map with attr of pre
func MapLang(attr, lang string) OptionFunc {
	return func(c Converter) error {
		if c.base() == nil {
			return errors.New("convert should contains a base converter")
		}
		c.base().langMap[attr] = lang
		return nil
	}
}
