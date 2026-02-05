package lang

import "testing"

func TestDefaultTargetLang(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want string
	}{
		{name: "ja", env: "ja_JP.UTF-8", want: "ja"},
		{name: "en", env: "en_US", want: "en"},
		{name: "zh", env: "zh_CN.UTF-8", want: "zh"},
		{name: "c", env: "C", want: "en"},
		{name: "empty", env: "", want: "en"},
		{name: "dash", env: "pt-BR", want: "pt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DefaultTargetLang(tc.env)
			if got != tc.want {
				t.Fatalf("DefaultTargetLang(%q) = %q, want %q", tc.env, got, tc.want)
			}
		})
	}
}
