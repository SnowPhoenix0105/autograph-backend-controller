package utils

import "unicode/utf8"

/*
SentenceSplitter 用于将长句子分割成若干较短句子， 这些句子的长度在 [minLen, maxLen] 区间内（包含上下界），尽量按照给定的分隔符进行分割。

对于一组分隔符，优先按照靠前的分隔符进行分割。
*/
type SentenceSplitter struct {

	/*
		index: 以byte为单位的下标
		cnt: 以rune为单位的下标
	*/

	// 输入
	separators []rune // 分隔符，按优先级降序排列
	maxLen     int    // 句子最大长度
	minLen     int    // 句子最小长度

	// 索引
	separatorCharToIndex map[rune]int // separators[i] -> i
	runeLenOfSeparator   []int        // separators[i] 在 utf8 下的编码字节数

	// 状态
	lastSplitIndex       int   // 上一次分割的 index
	lastSplitCnt         int   // 上一次分割的 cnt
	lastIndexOfSeparator []int // lastIndexOfSeparator[i] 表示 separators[i] 最后出现的 index
	lastCntOfSeparator   []int // lastIndexOfSeparator[i] 表示 separators[i] 最后出现的 index

	// 输出
	splitIndex []int // 分割句子的index
	splitCnt   []int // 分割句子的cnt
}

func (s *SentenceSplitter) updateLastIndexAndCntOfSeparator(ch rune, index, cnt int) {
	sep, ok := s.separatorCharToIndex[ch]
	if !ok {
		return
	}

	s.lastIndexOfSeparator[sep] = index
	s.lastCntOfSeparator[sep] = cnt
}

func (s *SentenceSplitter) addSplit(currentIndex, currentCnt int) {
	splitIndex := currentIndex
	splitCnt := currentCnt

	for i := 0; i < len(s.separators); i++ {
		cnt := s.lastCntOfSeparator[i] + 1
		if cnt >= s.lastSplitCnt+s.minLen {
			splitCnt = cnt
			splitIndex = s.lastIndexOfSeparator[i] + s.runeLenOfSeparator[i]
			break
		}
	}

	s.lastSplitCnt = splitCnt
	s.lastSplitIndex = splitIndex
	s.splitCnt = append(s.splitCnt, splitCnt)
	s.splitIndex = append(s.splitIndex, splitIndex)
}

func (s *SentenceSplitter) init() {
	s.separatorCharToIndex = make(map[rune]int, len(s.separators))
	s.runeLenOfSeparator = make([]int, len(s.separators))
	s.lastIndexOfSeparator = make([]int, len(s.separators))
	s.lastCntOfSeparator = make([]int, len(s.separators))

	for i, separator := range s.separators {
		s.separatorCharToIndex[separator] = i
		s.runeLenOfSeparator[i] = utf8.RuneLen(s.separators[i])
	}
}

func (s *SentenceSplitter) split(text string) {
	for i := 0; i < len(s.separators); i++ {
		s.lastIndexOfSeparator[i] = -1
		s.lastCntOfSeparator[i] = -1
	}

	cnt := -1
	for index, ch := range text {
		cnt += 1

		if cnt >= s.lastSplitCnt+s.maxLen {
			s.addSplit(index, cnt)
		}

		s.updateLastIndexAndCntOfSeparator(ch, index, cnt)
	}
}

/*
Split 将长句子分割成若干较短句子， 这些句子的长度在 [minLen, maxLen] 区间内（包含上下界），尽量按照给定的分隔符进行分割。

返回值：
	splitIndexOnByte: 作为utf8编码时，分割点的下标（不包括0和len(text)），对应 text[splitIndexOnByte[i] : splitIndexOnByte[i+1]]
	splitIndexOnRune: 作为unicode时，分割点的下标（不包括0和len([]rune(text))），对应 []rune(text)[splitIndexOnRune[i] : splitIndexOnRune[i+1]]
*/
func (s *SentenceSplitter) Split(text string) (splitIndexOnByte, splitIndexOnRune []int) {
	s.split(text)
	return s.splitIndex, s.splitCnt
}

/*
NewSentenceSplitter 构建一个SentenceSplitter。

SentenceSplitter 用于将长句子分割成若干较短句子， 这些句子的长度在 [minLen, maxLen] 区间内（包含上下界），尽量按照给定的分隔符进行分割。
*/
func NewSentenceSplitter(separators []rune, minLen, maxLen int) *SentenceSplitter {
	ret := SentenceSplitter{
		separators: separators,
		maxLen:     maxLen,
		minLen:     minLen,
	}
	ret.init()
	return &ret
}
