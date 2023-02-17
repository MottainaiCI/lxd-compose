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

func ParseCompressionMode(s string) CompressionMode {
	ans := None

	if s == "gz" || s == "gzip" {
		ans = Gzip
	} else if s == "zstd" {
		ans = Zstd
	} else if s == "xz" {
		ans = Xz
	} else if s == "bz2" || s == "bzip2" {
		ans = Bzip2
	}

	return ans
}

func NewTarCompressionOpts(useExt bool) *TarCompressionOpts {
	return &TarCompressionOpts{
		UseExt:         useExt,
		FileWriter:     nil,
		CompressWriter: nil,
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

func NewNopCloseWriter(buf *bufio.Writer) *NopCloseWriter {
	return &NopCloseWriter{Writer: buf}
}

func (ncw *NopCloseWriter) Close() error {
	ncw.Flush()
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
			if strings.HasSuffix(file, ".gz") || strings.HasSuffix(file, ".gzip") {
				cMode = Gzip
			} else if strings.HasSuffix(file, ".zstd") {
				cMode = Zstd
			} else if strings.HasSuffix(file, ".xz") {
				cMode = Xz
			} else if strings.HasSuffix(file, ".bz2") || strings.HasSuffix(file, ".bzip2") {
				cMode = Bzip2
			}
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
