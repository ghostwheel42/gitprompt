package gitprompt

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var all = &GitStatus{
	Branch:    "master",
	Sha:       "0455b83f923a40f0b485665c44aa068bc25029f5",
	Untracked: 0,
	Modified:  1,
	Staged:    2,
	Conflicts: 3,
	Ahead:     4,
	Behind:    5,
	Stashed:   6,
	Upstream:  "origin/master",
	Clean:     true,
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name     string
		status   *GitStatus
		format   string
		expected string
		width    int
	}{
		// output
		{
			name:     "empty",
			format:   "%e",
			expected: "",
		},
		{
			name:     "all data",
			format:   "%h %u %m %s %c %a %b %S %U",
			expected: "master 0 1 2 3 4 5 6 origin/master",
		},
		{
			name:     "unicode",
			format:   "%h ✋%u ⚡️%m 🚚%s ❗️%c ⬆%a ⬇%b",
			expected: "master ✋0 ⚡️1 🚚2 ❗️3 ⬆4 ⬇5",
			width:    26,
		},
		{
			name:     "sha",
			status:   &GitStatus{Sha: "858828b5e153f24644bc867598298b50f8223f9b"},
			format:   "%h%H",
			expected: "858828b:858828b",
		},
		// colors
		{
			name:     "red",
			format:   "#r%h",
			expected: "\x1b[31mmaster\x1b[0m",
			width:    6,
		},
		{
			name:     "bold",
			format:   "@b%h",
			expected: "\x1b[1mmaster\x1b[0m",
			width:    6,
		},
		{
			name:     "color & attribute",
			format:   "#r@bA",
			expected: "\x1b[1;31mA\x1b[0m",
			width:    1,
		},
		{
			name:     "color & attribute reversed",
			format:   "@b#rA",
			expected: "\x1b[1;31mA\x1b[0m",
			width:    1,
		},
		{
			name:     "ignore format until non-whitespace",
			format:   "A#r#g#b     B@i\tC",
			expected: "A     \x1b[34mB\t\x1b[3mC\x1b[0m",
			width:    9,
		},
		{
			name:     "reset color",
			format:   "#rA#_B",
			expected: "\x1b[31mA\x1b[0mB",
			width:    2,
		},
		{
			name:     "reset attributes",
			format:   "@bA@_B",
			expected: "\x1b[1mA\x1b[0mB",
			width:    2,
		},
		{
			name:     "reset attribute",
			format:   "#ggreen @b@igreen_bold_italic @Bgreen_italic",
			expected: "\x1b[32mgreen \x1b[1;3mgreen_bold_italic \x1b[0;3;32mgreen_italic\x1b[0m",
			width:    36,
		},
		{
			name:     "ending with #",
			format:   "%h#",
			expected: "master#",
		},
		{
			name:     "ending with !",
			format:   "%h!",
			expected: "master!",
		},
		{
			name:     "ending with @",
			format:   "%h@",
			expected: "master@",
		},
		// groups
		{
			name:     "groups",
			format:   "<[%h][ B%b A%a][ U%u][ C%c][ %CX][%ll][%eY]>",
			expected: "<master B5 A4 C3 XY>",
		},
		{
			name:     "group color auto-reset",
			format:   "<[#r%h]-[#g%u]%a[-#b%b]>",
			expected: "<\x1b[31mmaster\x1b[0m-4-\x1b[34m5\x1b[0m>",
			width:    12,
		},
		{
			name:     "group color leak",
			format:   "<[#r%h]-[#g%u]%a[-#b%b#>]>",
			expected: "<\x1b[31mmaster\x1b[0m-4-\x1b[34m5>\x1b[0m",
			width:    12,
		},
		{
			name:     "group attribute auto-reset",
			format:   "<[@b%h]-[@f%u]%a[-@i%b]>",
			expected: "<\x1b[1mmaster\x1b[0m-4-\x1b[3m5\x1b[0m>",
			width:    12,
		},
		{
			name:     "group attribute leak",
			format:   "<[@b%h]-[@f%u]%a[-@i%b@>]>",
			expected: "<\x1b[1mmaster\x1b[0m-4-\x1b[3m5>\x1b[0m",
			width:    12,
		},
		{
			name:     "group invalid",
			format:   "]",
			expected: "]",
		},
		// non matching
		{
			name:     "data valid odd",
			format:   "%%%h",
			expected: "%%master",
		},
		{
			name:     "data valid even",
			format:   "%%%%h",
			expected: "%%%%h",
		},
		{
			name:     "data invalid odd",
			format:   "%%%z",
			expected: "%%%z",
		},
		{
			name:     "data invalid even",
			format:   "%%%%z",
			expected: "%%%%z",
		},
		{
			name:     "color valid odd",
			format:   "###rA",
			expected: "##\x1b[31mA\x1b[0m",
			width:    3,
		},
		{
			name:     "color valid even",
			format:   "####rA",
			expected: "####rA",
		},
		{
			name:     "color invalid odd",
			format:   "###zA",
			expected: "###zA",
		},
		{
			name:     "color invalid even",
			format:   "####zA",
			expected: "####zA",
		},
		{
			name:     "attribute valid odd",
			format:   "@@@bA",
			expected: "@@\x1b[1mA\x1b[0m",
			width:    3,
		},
		{
			name:     "attribute valid even",
			format:   "@@@@bA",
			expected: "@@@@bA",
		},
		{
			name:     "attribute invalid odd",
			format:   "@@@zA",
			expected: "@@@zA",
		},
		{
			name:     "attribute invalid even",
			format:   "@@@@zA",
			expected: "@@@@zA",
		},
		{
			name:     "trailing %",
			format:   "A%",
			expected: "A%",
		},
		{
			name:     "trailing #",
			format:   "A#",
			expected: "A#",
		},
		{
			name:     "trailing @",
			format:   "A@",
			expected: "A@",
		},
		// escapes
		{
			name:     "data",
			format:   "A\\%h",
			expected: "A%h",
		},
		{
			name:     "color",
			format:   "A\\#rB",
			expected: "A#rB",
		},
		{
			name:     "attribute",
			format:   "A\\!bB",
			expected: "A!bB",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.status == nil {
				test.status = all
			}
			if test.width == 0 {
				test.width = len(test.expected)
			}
			actual := Print(test.status, test.format, true)

			// check zsh-formatted output: "%{<OUTPUT>%<WIDTH>G%}"
			expected := fmt.Sprintf("%%{%s%%%dG%%}", test.expected, test.width)
			if !strings.HasPrefix(actual, "%{") || !strings.HasSuffix(actual, "G%}") {
				fail(t, "Invalid zsh-formatted output", expected, actual)
				return
			}
			actual = actual[2 : len(actual)-3]

			split := strings.LastIndex(actual, "%")
			if split < 0 {
				fail(t, "Invalid zsh-formatted output", expected, actual)
				return
			}
			width, err := strconv.Atoi(actual[split+1:])
			actual = actual[:split]
			if err != nil {
				fail(t, "Invalid zsh-formatted output", expected, actual)
				return
			}

			if width != test.width {
				fail(t, fmt.Sprintf("Width does not match; expected %d, actual %d", test.width, width),
					test.expected, actual)
				return
			}

			if actual != test.expected {
				fail(t, "Output mismatch", test.expected, actual)
				return
			}

		})
	}
}

func fail(t *testing.T, message, expected, actual string) {
	t.Helper()
	t.Errorf(
		"%s\nExpected:  %s\n          %q\nActual:    %s\x1b[0m\n          %q",
		message, expected, expected, actual, actual,
	)
}
