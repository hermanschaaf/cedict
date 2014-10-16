// Copyright 2014 Herman Schaaf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package cedict provides a parser / tokenizer for reading entries from the CEDict
Chinese dictionary project.

Tokenizing is done by creating a CEDict for an io.Reader r. It is the
caller's responsibility to ensure that r provides a CEDict-formatted dictionary.

        import "github.com/hermanschaaf/cedict"

        ...

        c := cedict.New(r) // r is an io.Reader to the cedict file

Given a CEDict c, the dictionary is tokenized by repeatedly calling c.NextEntry(),
which parses until it reaches the next entry, or an error if no more entries are found:

        for {
            err := c.NextEntry()
            if err != nil {
                break
            }
            entry := c.Entry()
            fmt.Println(entry.Simplified, entry.Definitions[0])
        }

To retrieve the current entry, the Entry method can be called. There is also
a lower-level API available, using the bufio.Scanner Scan method. Using this
lower-level API is the recommended way to read comments from the CEDict, should
that be necessary.
*/
package cedict

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	EntryToken = iota
	CommentToken
	ErrorToken
)

// CEDict is the basic tokenizer struct we use to read and parse
// new dictionary instances.
type CEDict struct {
	*bufio.Scanner
	TokenType int
	entry     *Entry
}

// Entry represents a single entry in the cedict dictionary.
type Entry struct {
	Simplified      string
	Traditional     string
	Pinyin          string
	PinyinWithTones string
	PinyinNoTones   string
	Definitions     []string
}

// consumeComment reads from the data byte slice until a new line is found,
// returning the advanced steps, accumalated bytes and nil error if successful.
// This is done in accordance to the SplitFunc type defined in bufio.
func consumeComment(data []byte, atEOF bool) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == '\n' || (atEOF && i == len(data)-1) {
			return i + 1, accum, nil
		} else {
			accum = append(accum, b)
		}
	}
	if atEOF {
		return len(data), accum, nil
	}
	return 0, nil, nil
}

// consumeEntry reads from the data byte slice until a new line is found.
// It only returns the bytes found, and does not attempt to parse the actual
// entry on the line.
func consumeEntry(data []byte, atEOF bool) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == '\n' {
			return i + 1, accum, nil
		} else {
			accum = append(accum, b)
		}
	}
	if atEOF {
		return len(data), accum, nil
	}
	return 0, nil, nil
}

// New takes an io.Reader and creates a new CEDict instance.
func New(r io.Reader) *CEDict {
	s := bufio.NewScanner(r)
	c := &CEDict{
		Scanner: s,
	}
	// splitFunc defines how we split our tokens
	splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if data[0] == '#' {
			advance, token, err = consumeComment(data, atEOF)
			c.TokenType = CommentToken
		} else {
			advance, token, err = consumeEntry(data, atEOF)
			c.TokenType = EntryToken
		}
		return
	}
	s.Split(splitFunc)
	return c
}

// toneLookupTable returns a lookup table to replace a specified tone number with
// its appropriate UTF-8 character with tone marks
func toneLookupTable(tone int) (map[string]string, error) {
	if tone < 0 || tone > 5 {
		return nil, fmt.Errorf("Tried to create tone lookup table with tone %i", tone)
	}

	lookupTable := map[string][]string{
		"a": []string{"a", "ā", "á", "ǎ", "à", "a"},
		"e": []string{"e", "ē", "é", "ě", "è", "e"},
		"i": []string{"i", "ī", "í", "ǐ", "ì", "i"},
		"o": []string{"o", "ō", "ó", "ǒ", "ò", "o"},
		"u": []string{"u", "ū", "ú", "ǔ", "ù", "u"},
		"ü": []string{"ü", "ǖ", "ǘ", "ǚ", "ǜ", "ü"},
	}

	toneLookup := make(map[string]string)

	for vowel, toneRunes := range lookupTable {
		toneLookup[vowel] = toneRunes[tone]
	}

	return toneLookup, nil
}

// extractTone splits the tone number and the pinyin syllable returning a string
// and an integer, e.g., dong1 => dong, 1
func extractTone(p string) (string, int) {
	tone := int(p[len(p)-1]) - 48

	if tone < 0 || tone > 5 {
		return p, 0
	}
	return p[0 : len(p)-1], tone
}

// replaceWithToneMark returns the UTF-8 representation of a pinyin syllable with
// the appropriate tone, e.g., dong1 => dōng, using the pinyin accent placement rules
func replaceWithToneMark(s string, tone int) (string, error) {
	lookup, err := toneLookupTable(tone)
	if err != nil {
		return "", err
	}

	if strings.Contains(s, "a") {
		return strings.Replace(s, "a", lookup["a"], -1), nil
	}
	if strings.Contains(s, "e") {
		return strings.Replace(s, "e", lookup["e"], -1), nil
	}
	if strings.Contains(s, "ou") {
		return strings.Replace(s, "o", lookup["o"], -1), nil
	}
	index := strings.LastIndexAny(s, "iüou")
	if index != -1 {
		var out bytes.Buffer
		for ind, runeValue := range s {
			if ind == index {
				out.WriteString(lookup[string(runeValue)])
			} else {
				out.WriteString(string(runeValue))
			}
		}
		return out.String(), nil
	}
	return "", fmt.Errorf("No tone match")
}

// convertToTones takes a CEDICT pinyin representation and returns the concatenated
// pinyin version with tone marks, e.g., yi1 lan3 zi5 => yīlǎnzi
func convertToTones(p string) string {
	pv := strings.Replace(p, "u:", "ü", -1)
	py := strings.Split(pv, " ")

	var out bytes.Buffer
	for _, pySyllable := range py {
		pyNoTone, tone := extractTone(pySyllable)
		pyWithTone, err := replaceWithToneMark(pyNoTone, tone)
		if err != nil {
			return ""
		}
		out.WriteString(pyWithTone)
	}
	return out.String()
}

// pinyinNoTones takes a CEDICT pinyin representation and returns the concatenated
// pinyin version without tone marks, e.g., yi1 lan3 zi5 => yilanzi
// This representation is useful for building a search interface to the CEDICT database
// for user pinyin input.
// Note: This substitutes the more common search term "v" for "ü"
func pinyinNoTones(p string) string {
	pv := strings.Replace(p, "u:", "v", -1)
	py := strings.Split(pv, " ")

	var out bytes.Buffer
	for _, pySyllable := range py {
		pyNoTone, _ := extractTone(pySyllable)
		out.WriteString(pyNoTone)
	}
	return out.String()
}

var reEntry = regexp.MustCompile(`(?P<trad>\S*?) (?P<simp>\S*?) \[(?P<pinyin>.+)\] \/(?P<defs>.+)\/`)

// parseEntry parses string entries from CEDict of the form:
//     一之為甚 一之为甚 [yi1 zhi1 wei2 shen4] /Once is enough (idiom)/
// It returns a pointer to a new Entry struct.
func parseEntry(s string) (*Entry, error) {
	match := reEntry.FindStringSubmatch(s)
	if match == nil {
		return nil, fmt.Errorf("Badly formatted entry: %v", s)
	}

	e := Entry{}
	for i, name := range reEntry.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}
		switch name {
		case "simp":
			e.Simplified = match[i]
		case "trad":
			e.Traditional = match[i]
		case "pinyin":
			e.Pinyin = strings.ToLower(match[i])
		case "defs":
			e.Definitions = strings.Split(match[i], "/")
		}
	}
	e.PinyinWithTones = convertToTones(e.Pinyin)
	e.PinyinNoTones = pinyinNoTones(e.Pinyin)
	return &e, nil
}

var NoMoreEntries error = errors.New("No more entries to read")

// Next reads until the next entry token is found. Once found,
// it parses the token and returns a pointer to a newly populated
// Entry struct.
func (c *CEDict) NextEntry() error {
	for c.Scan() {
		if c.TokenType == EntryToken {
			e, err := parseEntry(c.Text())
			if err != nil {
				return err
			}
			c.entry = e
			return nil
		}
	}
	if err := c.Err(); err != nil {
		return err
	}

	return NoMoreEntries
}

// Entry returns a pointer to the most recently parsed Entry struct.
func (c *CEDict) Entry() *Entry {
	return c.entry
}
