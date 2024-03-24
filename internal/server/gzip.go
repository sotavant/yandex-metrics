package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		fmt.Println("lajer")
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.zr.Close()
}
