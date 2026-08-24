package guard_test

import (
	"reflect"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestPreprocessLangs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "nil input returns nil",
			in:   nil,
			want: nil,
		},
		{
			name: "empty input returns nil",
			in:   []string{},
			want: nil,
		},
		{
			name: "single language",
			in:   []string{"en"},
			want: []string{"en"},
		},
		{
			name: "comma separated languages",
			in:   []string{"en,fr,lv"},
			want: []string{"en", "fr", "lv"},
		},
		{
			name: "trims spaces",
			in:   []string{" en , fr , lv "},
			want: []string{"en", "fr", "lv"},
		},
		{
			name: "drops empty parts",
			in:   []string{"en,, fr, ,lv,"},
			want: []string{"en", "fr", "lv"},
		},
		{
			name: "deduplicates languages",
			in:   []string{"en,fr", "fr", "lv,en"},
			want: []string{"en", "fr", "lv"},
		},
		{
			name: "sorts result",
			in:   []string{"lv,en,fr"},
			want: []string{"en", "fr", "lv"},
		},
		{
			name: "preserves case",
			in:   []string{"EN,en"},
			want: []string{"EN", "en"},
		},
		{
			name: "preserves hyphenated locale tags",
			in:   []string{"pt-BR, en-US"},
			want: []string{"en-US", "pt-BR"},
		},
		{
			name: "only empty strings returns empty slice",
			in:   []string{"", " , ,, "},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := guard.PreprocessLangs(tt.in)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf(
					"PreprocessLangs(%#v) = %#v, want %#v",
					tt.in,
					got,
					tt.want,
				)
			}
		})
	}
}
