/*
Copyright (C) 2021-2023  Daniele Rondina <geaaru@gmail.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package tools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	bzip2 "github.com/dsnet/compress/bzip2"
	zstd "github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	xz "github.com/ulikunitz/xz"
)

type CompressionMode string

const (
	None  CompressionMode = "none"
	Gzip  CompressionMode = "gz"
	Zstd  CompressionMode = "zstd"
	Xz    CompressionMode = "xz"
	Bzip2 CompressionMode = "bz2"
)

type TarCompressionOpts struct {
	UseExt         bool
	Mode           CompressionMode
	FileWriter     io.WriteCloser
	CompressWriter io.WriteCloser
}

type TarReaderCompressionOpts struct {
	UseExt         bool
	Mode           CompressionMode
	CompressReader io.ReadCloser
	FileReader     io.ReadCloser
}

func ParseCompressionMode(s string) CompressionMode {
	ans := None

	if s == "gz" || s == "gzip" {
		ans = Gzip
	} else if s == "zstd" || s == "zst" {
		ans = Zstd
	} else if s == "xz" {
		ans = Xz
	} else if s == "bz2" || s == "bzip2" {
		ans = Bzip2
	}

	return ans
}

func GetCompressionMode(file string) CompressionMode {
	cMode := None

	if strings.HasSuffix(file, ".gz") || strings.HasSuffix(file, ".gzip") {
		cMode = Gzip
	} else if strings.HasSuffix(file, ".zstd") || strings.HasSuffix(file, ".zst") {
		cMode = Zstd
	} else if strings.HasSuffix(file, ".xz") {
		cMode = Xz
	} else if strings.HasSuffix(file, ".bz2") || strings.HasSuffix(file, ".bzip2") {
		cMode = Bzip2
	}
	return cMode
}

func NewTarCompressionOpts(useExt bool) *TarCompressionOpts {
	return &TarCompressionOpts{
		UseExt:         useExt,
		FileWriter:     nil,
		CompressWriter: nil,
	}
}

func NewTarReaderCompressionOpts(useExt bool) *TarReaderCompressionOpts {
	return &TarReaderCompressionOpts{
		UseExt:         useExt,
		FileReader:     nil,
		CompressReader: nil,
	}
}

func (o *TarReaderCompressionOpts) Close() {
	if o.CompressReader != nil {
		o.CompressReader.Close()
		o.CompressReader = nil
	}

	if o.FileReader != nil {
		o.FileReader.Close()
		o.CompressReader = nil
	}
}

func (o *TarCompressionOpts) Close() {
	if o.CompressWriter != nil {
		o.CompressWriter.Close()
		o.CompressWriter = nil
	}

	if o.FileWriter != nil {
		o.FileWriter.Close()
		o.FileWriter = nil
	}
}

type NopCloseWriter struct {
	*bufio.Writer
}

type NopCloseReader struct {
	io.Reader
}

func NewNopCloseWriter(buf *bufio.Writer) *NopCloseWriter {
	return &NopCloseWriter{Writer: buf}
}

func (ncw *NopCloseWriter) Close() error {
	ncw.Flush()
	return nil
}

func NewNopCloseReader(r io.Reader) *NopCloseReader {
	return &NopCloseReader{Reader: r}
}

func (ncw *NopCloseReader) Close() error {
	return nil
}

func PrepareTarReader(file string, opts *TarReaderCompressionOpts) error {
	var err error
	cMode := None

	if file == "-" {
		// POST: Using stdint for read
		r := bufio.NewReader(os.Stdin)
		opts.FileReader = NewNopCloseReader(r)

		if !opts.UseExt {
			cMode = opts.Mode
		}
	} else {
		opts.FileReader, err = os.OpenFile(file, os.O_RDONLY, 0666)
		if err != nil {
			return fmt.Errorf(
				"Error on open file %s: %s", file, err.Error())
		}

		if opts.UseExt {
			cMode = GetCompressionMode(file)
		} else {
			cMode = opts.Mode
		}
	}

	switch cMode {
	case Gzip:
		opts.CompressReader, err = gzip.NewReader(opts.FileReader)
		if err != nil {
			return err
		}
	case Zstd:
		r, err := zstd.NewReader(opts.FileReader)
		if err != nil {
			return err
		}
		opts.CompressReader = r.IOReadCloser()
	case Xz:
		r, err := xz.NewReader(opts.FileReader)
		if err != nil {
			return err
		}
		opts.CompressReader = NewNopCloseReader(r)
	case Bzip2:
		opts.CompressReader, err = bzip2.NewReader(opts.FileReader, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func PrepareTarWriter(file string, opts *TarCompressionOpts) error {
	var err error
	cMode := None

	if file == "-" {
		// POST: Using stdout for write
		w := bufio.NewWriter(os.Stdout)
		opts.FileWriter = NewNopCloseWriter(w)

		if !opts.UseExt {
			cMode = opts.Mode
		}
	} else {
		opts.FileWriter, err = os.Create(file)
		if err != nil {
			return fmt.Errorf(
				"Error on create file %s: %s", file, err.Error())
		}

		if opts.UseExt {
			cMode = GetCompressionMode(file)
		} else {
			cMode = opts.Mode
		}
	}

	switch cMode {

	case Gzip:
		w := gzip.NewWriter(opts.FileWriter)
		w.SetConcurrency(1<<20, runtime.NumCPU())
		opts.CompressWriter = w
	case Zstd:
		opts.CompressWriter, err = zstd.NewWriter(opts.FileWriter)
		if err != nil {
			return err
		}
	case Xz:
		opts.CompressWriter, err = xz.NewWriter(opts.FileWriter)
		if err != nil {
			return err
		}
	case Bzip2:
		opts.CompressWriter, err = bzip2.NewWriter(opts.FileWriter, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
