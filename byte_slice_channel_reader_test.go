package fragmentedbuf

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"math/rand"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type ByteSliceChannelReaderSuite struct {
	suite.Suite
	ch   chan []byte
	bscr *ByteSliceChannelReader
}

func (s *ByteSliceChannelReaderSuite) SetupSuite() {
	s.ch = make(chan []byte, 5)
	s.bscr = NewByteSliceChannelReader(s.ch)
}

func test(t *testing.T, label string, writeSize int, readSize int) {
	require := require.New(t)

	ch := make(chan []byte, 5)
	bscr := NewByteSliceChannelReader(ch)

	expected := randomString(1147)

	written := 0
	input := bytes.NewBuffer(nil)
	input.WriteString(expected)

	go func() {
		defer close(ch)
		for written < len(expected) {
			remaining := len(expected) - written
			chunkSize := remaining
			if chunkSize > writeSize {
				chunkSize = writeSize
			}
			writeChunk := make([]byte, writeSize)
			n, err := input.Read(writeChunk[:chunkSize])
			require.Nil(err, label)
			written += n
			ch <- writeChunk[:chunkSize]
		}
	}()

	read := 0
	got := bytes.NewBuffer(nil)
	buf := make([]byte, readSize)
	for read < len(expected) {
		remaining := len(expected) - got.Len()
		chunkSize := remaining
		if chunkSize > readSize {
			chunkSize = readSize
		}
		n, err := bscr.Read(buf[:chunkSize])
		require.Nil(err, label)
		read += n
		got.Write(buf[:chunkSize])
	}
	require.Equal(expected, got.String())
}

func (s *ByteSliceChannelReaderSuite) TestBasic() {
	type entry struct {
		writeSize int
		readSize  int
	}

	largeReaderSmallWriterTable := []entry{
		{1, 1},
		{1, 2},
		{1, 6},
		{1, 8},
		{5, 8},
		{5, 16},
		{7, 16},
		{8, 16},
		{9, 16},
		{7, 24},
		{8, 24},
		{9, 24},
		{8, 24},
		{8, 32},
		{8, 64},
		{8, 63},
		{8, 65},
		{32, 127},
		{32, 127},
	}

	smallReaderLargeWriterTable := make([]entry, 0)
	// Reverse the values
	for _, v := range largeReaderSmallWriterTable {
		n := entry{
			writeSize: v.readSize,
			readSize:  v.writeSize,
		}
		smallReaderLargeWriterTable = append(smallReaderLargeWriterTable, n)
	}

	finalTests := make([]entry, 0)
	finalTests = append(finalTests, largeReaderSmallWriterTable...)
	finalTests = append(finalTests, smallReaderLargeWriterTable...)

	wg := sync.WaitGroup{}
	wg.Add(len(finalTests))

	for _, v := range finalTests {
		go func(v *entry) {
			defer wg.Done()
			label := fmt.Sprintf("write:%v read:%v", v.writeSize, v.readSize)
			test(s.T(), label, v.writeSize, v.readSize)
		}(&v)
	}
	wg.Wait()
}

func TestByteSliceChannelReader(t *testing.T) {
	suite.Run(t, new(ByteSliceChannelReaderSuite))
}
