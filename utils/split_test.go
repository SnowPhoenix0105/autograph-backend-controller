package utils

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStringRange(t *testing.T) {
	for i, ch := range "abc中文" {
		t.Logf("[%d]%c", i, ch)
	}
	t.Log(len("中"))
}

func assertSplitterResultValid(t *testing.T, s *SentenceSplitter, text string) {
	chars := []rune(text)
	cntToStr := func(beg, end int) string {
		b := strings.Builder{}
		for i := beg; i < end; i++ {
			b.WriteRune(chars[i])
		}
		return b.String()
	}

	cntStr := cntToStr(0, s.splitCnt[0])
	t.Log(cntStr)
	require.Equal(t, cntStr, text[0:s.splitIndex[0]])

	for i := 1; i < len(s.splitIndex); i++ {
		cntStr = cntToStr(s.splitCnt[i-1], s.splitCnt[i])
		t.Log(cntStr)
		require.Equal(t, cntStr, text[s.splitIndex[i-1]:s.splitIndex[i]])
	}

	cntStr = cntToStr(s.splitCnt[len(s.splitCnt)-1], utf8.RuneCountInString(text))
	t.Log(cntStr)
	require.Equal(t, cntStr, text[s.splitIndex[len(s.splitIndex)-1]:])
}

func TestSplitter_SplitWithNoSepText(t *testing.T) {
	s := SentenceSplitter{
		separators: []rune{'。', '，', '；', ',', '.'},
		maxLen:     2,
		minLen:     1,
	}
	s.init()
	text := "abcd一二三四五"
	s.split(text)

	assertSplitterResultValid(t, &s, text)

	require.Equal(t, []int{2, 4, 6, 8}, s.splitCnt)
}

func TestSplitter_SplitWithFullSepText(t *testing.T) {
	s := SentenceSplitter{
		separators: []rune{'。', '，', '；', ',', '.'},
		maxLen:     2,
		minLen:     1,
	}
	s.init()
	text := ",,,,。。，，；"
	s.split(text)

	assertSplitterResultValid(t, &s, text)

	require.Equal(t, []int{2, 4, 6, 8}, s.splitCnt)
}

func TestSplitter_SplitSepPriority(t *testing.T) {
	s := SentenceSplitter{
		separators: []rune{'。', '，', '；', ',', '.'},
		maxLen:     8,
		minLen:     4,
	}
	s.init()
	text := "。。。。，；,.ooo。o"
	s.split(text)

	assertSplitterResultValid(t, &s, text)

	require.Equal(t, []int{4, 12}, s.splitCnt)
}
