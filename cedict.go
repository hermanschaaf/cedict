// Copyright 2014 Herman Schaaf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package cedict provides a parser / tokenizer for reading entries from the CEDict
Chinese dictionary project.

Tokenizing is done by creating a CEDict for an io.Reader r. It is the
caller's responsibility to ensure that r provides a CEDict-formatted dictionary.

        c := cedict.New(r)

Given a CEDict c, the dictionary is tokenized by repeatedly calling c.Next(),
which parses the next token and returns its contents as an Entry, or an error:

        for {
                entry, err := c.Next()
                if err != nil {
                        // handle error...
                        return ...
                }
                // Process the current entry.
        }

To retrieve the current entry, the Entry method can be called. There is also
a lower-level API available, using the bufio.Scanner Scan method. Using this
API is the recommended way to read comments from the CEDict, should that
be necessary.
*/
package cedict

import (
	"bufio"
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
	Entry     *Entry
}

// Entry represents a single entry in the cedict dictionary.
type Entry struct {
	Simplified  string
	Traditional string
	Pinyin      string
	Definitions []string
}

// consumeComment reads from the data byte slice until a new line is found,
// returning the advanced steps, accumalated bytes and nil error if successful.
// This is done in accordance to the SplitFunc type defined in bufio.
func consumeComment(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == '\n' {
			return i, accum, nil
		} else {
			accum = append(accum, b)
		}
	}
	return 0, nil, nil
}

// consumeEntry reads from the data byte slice until a new line is found.
// It only returns the bytes found, and does not attempt to parse the actual
// entry on the line.
func consumeEntry(data []byte) (int, []byte, error) {
	var accum []byte
	for i, b := range data {
		if b == '\n' {
			return i, accum, nil
		} else {
			accum = append(accum, b)
		}
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
			advance, token, err = consumeComment(data)
			c.TokenType = CommentToken
		} else {
			advance, token, err = consumeEntry(data)
			c.TokenType = EntryToken
		}
		return
	}
	s.Split(splitFunc)
	return c
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
			e.Pinyin = match[i]
		case "defs":
			e.Definitions = strings.Split(match[i], "/")
		}
	}
	return &e, nil
}

type NoMoreEntries error

// Next reads until the next entry token is found. Once found,
// it parses the token and returns a pointer to a newly populated
// Entry struct.
func (c *CEDict) Next() error {
	for c.Scan() {
		if c.TokenType == EntryToken {
			e, err := parseEntry(c.Text())
			if err != nil {
				return err
			}
			c.Entry = e
			return nil
		}
	}
	return NoMoreEntries(errors.New("No more entries to read"))
}
