package sanitize

import (
	"strings"
	"testing"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"printable ascii passes through", "hello world", "hello world"},
		{"strips control chars", "hello\x00\x01\x02world", "helloworld"},
		{"strips escape sequences", "hello\x1b[31mworld", "hello[31mworld"},
		{"strips null bytes", "he\x00llo", "hello"},
		{"preserves space", "foo bar", "foo bar"},
		{"preserves punctuation", "!@#$%^&*()", "!@#$%^&*()"},
		{"strips tabs and newlines", "hello\t\nworld", "helloworld"},
		{"empty string", "", ""},
		// Emoji support
		{"preserves emoji", "hello 🍺 world", "hello 🍺 world"},
		{"preserves multiple emojis", "🎵🎶🎤", "🎵🎶🎤"},
		{"preserves emoji in sentence", "cheers 🍻 to that!", "cheers 🍻 to that!"},
		{"preserves compound emoji", "👋🏽", "👋🏽"},
		{"preserves common emojis", "❤️🔥✨💀👀🎉", "❤️🔥✨💀👀🎉"},
		// International characters
		{"preserves accented chars", "café résumé", "café résumé"},
		{"preserves CJK", "你好世界", "你好世界"},
		{"preserves japanese", "こんにちは", "こんにちは"},
		{"preserves korean", "안녕하세요", "안녕하세요"},
		{"preserves cyrillic", "привет", "привет"},
		// Mixed
		{"mixed ascii and emoji", "hello 🌍! 你好", "hello 🌍! 你好"},
		{"strips control but keeps emoji", "\x00🍺\x01hello\x02🎵", "🍺hello🎵"},
		// Invalid UTF-8
		{"strips invalid utf8", "hello\x80\xff", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clean(tt.input)
			if got != tt.want {
				t.Errorf("Clean(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanNick(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid nick", "dusty_dev", "dusty_dev", false},
		{"strips bad chars", "dust\x00y", "dusty", false},
		{"too short", "a", "", true},
		{"too long", "abcdefghijklmnopqrstu", "", true},
		{"min length", "ab", "ab", false},
		{"max length", "abcdefghijklmnopqrst", "abcdefghijklmnopqrst", false},
		{"empty after clean", "\x00\x01", "", true},
		// Nicknames are ASCII-only
		{"strips emoji from nick", "cool🍺guy", "coolguy", false},
		{"strips unicode from nick", "café", "caf", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanNick(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanNick(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CleanNick(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanChat(t *testing.T) {
	t.Run("normal message", func(t *testing.T) {
		got := CleanChat("hello everyone")
		if got != "hello everyone" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("preserves emoji", func(t *testing.T) {
		got := CleanChat("cheers 🍻!")
		if got != "cheers 🍻!" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("truncates at 500 runes not bytes", func(t *testing.T) {
		// 499 'a' + 1 emoji = 500 runes
		input := strings.Repeat("a", 499) + "🍺"
		got := CleanChat(input)
		runes := []rune(got)
		if len(runes) != 500 {
			t.Errorf("rune count = %d, want 500", len(runes))
		}
		if runes[499] != '🍺' {
			t.Errorf("last rune = %c, want 🍺", runes[499])
		}
	})
	t.Run("truncates cleanly at rune boundary", func(t *testing.T) {
		// 501 emojis — should truncate to 500
		input := strings.Repeat("🍺", 501)
		got := CleanChat(input)
		runes := []rune(got)
		if len(runes) != 500 {
			t.Errorf("rune count = %d, want 500", len(runes))
		}
	})
	t.Run("strips escape codes", func(t *testing.T) {
		got := CleanChat("\x1b[0mhello")
		if got != "[0mhello" {
			t.Errorf("got %q", got)
		}
	})
}

func TestCleanNote(t *testing.T) {
	t.Run("preserves emoji", func(t *testing.T) {
		got := CleanNote("🍺 cheers!")
		if got != "🍺 cheers!" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("truncates at 280 runes", func(t *testing.T) {
		input := strings.Repeat("🍺", 300)
		got := CleanNote(input)
		runes := []rune(got)
		if len(runes) != 280 {
			t.Errorf("rune count = %d, want 280", len(runes))
		}
	})
	t.Run("strips control chars", func(t *testing.T) {
		got := CleanNote("hello\x00\x01world")
		if got != "helloworld" {
			t.Errorf("got %q", got)
		}
	})
}
