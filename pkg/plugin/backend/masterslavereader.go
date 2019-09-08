package backend

import (
	"bytes"
	"io"
)

type (
	// masterSlaveReader reads bytes to master,
	// and synchronize them to slave.
	// Currently only support one slave.
	masterSlaveReader struct {
		masterReader io.Reader
		slaveReader  io.Reader
	}

	masterReader struct {
		r        io.Reader
		buffChan chan []byte
	}

	slaveReader struct {
		unreadBuff *bytes.Buffer
		buffChan   chan []byte
	}
)

func newMasterSlaveReader(r io.Reader) (io.Reader, io.Reader) {
	buffChan := make(chan []byte, 10)
	mr := &masterReader{
		r:        r,
		buffChan: buffChan,
	}
	sr := &slaveReader{
		unreadBuff: bytes.NewBuffer(nil),
		buffChan:   buffChan,
	}

	return mr, sr
}

func (mr *masterReader) Read(p []byte) (n int, err error) {
	buff := bytes.NewBuffer(nil)
	tee := io.TeeReader(mr.r, buff)
	n, err = tee.Read(p)

	if n != 0 {
		mr.buffChan <- buff.Bytes()
	}

	if err == io.EOF {
		close(mr.buffChan)
	}

	return n, err
}

func (sr *slaveReader) Read(p []byte) (int, error) {
	buff, ok := <-sr.buffChan

	if !ok {
		return 0, io.EOF
	}

	var n int
	// NOTE: This if-branch is defensive programming,
	// Because the callers of Read of both master and slave
	// are the same, so it never happens that len(p) < len(buff).
	// else-branch is faster because it is one less copy operation than if-branch.
	if sr.unreadBuff.Len() > 0 || len(p) < len(buff) {
		sr.unreadBuff.Write(buff)
		n, _ = sr.unreadBuff.Read(p)
	} else {
		n = copy(p, buff)
	}

	return n, nil
}