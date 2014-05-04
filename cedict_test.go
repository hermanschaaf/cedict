// Copyright 2014 Herman Schaaf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cedict

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// ExampleCEDict demonstrates basic usage of the cedict package. It uses a string.Reader as
// io.Reader, where normally you would use a file.Reader. Other than that, this demonstrates
// typical usage.
func ExampleCEDict() {
	dict := `一層 一层 [yi1 ceng2] /layer/
一攬子 一揽子 [yi1 lan3 zi5] /all-inclusive/undiscriminating/`
	r := io.Reader(strings.NewReader(dict))
	c := New(r)
	for {
		err := c.NextEntry()
		if err != nil {
			// you may also compare the error to cedict.NoMoreEntries
			// to know whether the end was reached or some other problem
			// occurred.
			break
		}
		// get current entry
		entry := c.Entry()
		// print out some fields
		fmt.Println(entry.Simplified, entry.Definitions[0])
	}
	// Output:
	// 一层 layer
	// 一揽子 all-inclusive
}

// TestParseEntry tests the parsing of individual entries, and checks that
// the correct fields are entered into the Entry struct.
func TestParseEntry(t *testing.T) {
	tests := []struct {
		give string
		want Entry
	}{
		{
			give: "一之為甚 一之为甚 [yi1 zhi1 wei2 shen4] /Once is enough (idiom)/",
			want: Entry{
				Simplified:  "一之为甚",
				Traditional: "一之為甚",
				Pinyin:      "yi1 zhi1 wei2 shen4",
				Definitions: []string{"Once is enough (idiom)"},
			},
		},
		{
			give: "一壁 一壁 [yi1 bi4] /one side/at the same time/",
			want: Entry{
				Simplified:  "一壁",
				Traditional: "一壁",
				Pinyin:      "yi1 bi4",
				Definitions: []string{"one side", "at the same time"},
			},
		},
	}
	for _, tt := range tests {
		got, err := parseEntry(tt.give)
		if err != nil {
			t.Fatalf("parseEntry(%q) error: %v", tt.give, err)
		}
		if !reflect.DeepEqual(*got, tt.want) {
			t.Errorf("parseEntry(%q):\ngot\t%v,\nwant\t%v", tt.give, *got, tt.want)
		}
	}
}

// TestCEDict parses a simple CEDict and puts the parsed values into Entry structs.
// It then confirms that these structs match our expectations.
func TestCEDict(t *testing.T) {
	raw := `# CC-CEDICT
# Community maintained free Chinese-English dictionary.
一團火 一团火 [yi1 tuan2 huo3] /fireball/ball of fire/
一團 一团 [yi1 tuan2] /1 regiment/
一層 一层 [yi1 ceng2] /layer/
一攬子 一揽子 [yi1 lan3 zi5] /all-inclusive/undiscriminating/
一東一西 一东一西 [yi1 dong1 yi1 xi1] /far apart/`
	want := []Entry{
		{Simplified: "一团火", Traditional: "一團火", Pinyin: "yi1 tuan2 huo3", Definitions: []string{"fireball", "ball of fire"}},
		{Simplified: "一团", Traditional: "一團", Pinyin: "yi1 tuan2", Definitions: []string{"1 regiment"}},
		{Simplified: "一层", Traditional: "一層", Pinyin: "yi1 ceng2", Definitions: []string{"layer"}},
		{Simplified: "一揽子", Traditional: "一攬子", Pinyin: "yi1 lan3 zi5", Definitions: []string{"all-inclusive", "undiscriminating"}},
		{Simplified: "一东一西", Traditional: "一東一西", Pinyin: "yi1 dong1 yi1 xi1", Definitions: []string{"far apart"}},
	}
	r := io.Reader(strings.NewReader(raw))
	c := New(r)
	entries := []Entry{}
	for {
		err := c.NextEntry()
		if err == NoMoreEntries {
			break
		} else if err != nil {
			t.Fatalf("CEDict.NextEntry() error: %v", err)
			break
		}
		// Process the current entry:
		e := c.Entry()
		entries = append(entries, *e)
	}
	if len(entries) != 5 {
		t.Fatalf("len(entries): got %d, want %d", len(entries), 5)
	}
	for i := range entries {
		if !reflect.DeepEqual(entries[i], want[i]) {
			t.Errorf("CEDict.Entry():\ngot\t%v,\nwant\t%v", entries[i], want[i])
		}
	}
}
