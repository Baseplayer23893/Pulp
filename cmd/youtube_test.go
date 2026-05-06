package cmd

import (
	"strings"
	"testing"
)

func TestParseVTTHandlesMinuteTimestamps(t *testing.T) {
	vtt := strings.Join([]string{
		"WEBVTT",
		"",
		"00:05.000 --> 00:07.000",
		"hello world",
		"",
		"00:08,000 --> 00:10,000",
		"next line",
	}, "\n")

	got := parseVTT(vtt)
	want := "hello world next line"
	if got != want {
		t.Fatalf("parseVTT() = %q, want %q", got, want)
	}
}
