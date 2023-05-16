package fragmentedbuf

import "io"

type ByteSliceChannelReader struct {
	c   <-chan []byte
	buf []byte
}

// NewByteSliceChannelReader creates a new instance of ByteSliceChannelReader.
// The boolean flag (shouldCopy) indicates whether the byte slices should be copied or referenced directly.
func NewByteSliceChannelReader(c <-chan []byte) *ByteSliceChannelReader {
	return &ByteSliceChannelReader{
		c:   c,
		buf: nil,
	}
}

func (bscr *ByteSliceChannelReader) Read(buf []byte) (int, error) {
	read := 0
	for {
		if bscr.buf != nil {
			n := copy(buf, bscr.buf)
			read += n
			bscr.buf = bscr.buf[n:]
			if len(bscr.buf) == 0 {
				bscr.buf = nil
			}
			if read == len(buf) {
				return read, nil
			}
		}
		// If we got here, then bscr.buf is nil (either was nil to begin with or has been exhausted) and we need to pull in new data
		incomingData, ok := <-bscr.c
		if incomingData != nil {
			// Copy over what we can
			n := copy(buf[read:], incomingData)
			read += n
			// Update the internal buffer to point to the remaining data
			bscr.buf = make([]byte, len(incomingData[n:]))
			copy(bscr.buf, incomingData[n:])
			if read == len(buf) {
				return read, nil
			}
		}
		if !ok {
			return read, io.EOF
		}
	}
}
