package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipWriter struct {
	rw http.ResponseWriter
	zw *gzip.Writer
}

func NewGzipWriter(rw http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		rw: rw,
		zw: gzip.NewWriter(rw),
	}
}

func (gw *GzipWriter) Header() http.Header {
	return gw.rw.Header()
}

func (gw *GzipWriter) Write(bytes []byte) (int, error) {
	return gw.zw.Write(bytes)
}

func (gw *GzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.rw.Header().Set("Content-Encoding", "gzip")
	}
	gw.rw.WriteHeader(statusCode)
}

func (gw *GzipWriter) Close() error {
	return gw.zw.Close()
}

type GzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &GzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (gr *GzipReader) Read(bytes []byte) (int, error) {
	return gr.zr.Read(bytes)
}

func (gr *GzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}
