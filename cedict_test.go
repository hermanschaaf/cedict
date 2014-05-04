package cedict

import (
	"reflect"
	"testing"
)

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
			t.Errorf("parseEntry(%q): got %+v, want %+v", tt.give, *got, tt.want)
		}
	}
}
