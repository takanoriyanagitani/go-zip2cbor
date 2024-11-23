package zipstd

import (
	"archive/zip"
	"context"
	"io"
	"iter"
	"os"

	zc "github.com/takanoriyanagitani/go-zip2cbor"
	util "github.com/takanoriyanagitani/go-zip2cbor/util"

	bo "github.com/takanoriyanagitani/go-zip2cbor/basic/out"
)

// Max number of bytes of the item of the zip file.
const MaxZipItemBytesDefault int64 = 16777216

type ZipToBasicHeader func(*zip.FileHeader, *zc.BasicZipFileHeader)

type StdMethodToBasicMethodMap map[uint16]zc.CompressionMethod

var StdMethodToBasicMethodMapDefault StdMethodToBasicMethodMap = map[uint16]zc.
	CompressionMethod{
	zip.Store:   zc.CompressionMethodStore,
	zip.Deflate: zc.CompressionMethodDeflate,
}

func MethodConvert(std uint16) zc.CompressionMethod {
	basic, found := StdMethodToBasicMethodMapDefault[std]
	switch found {
	case true:
		return basic
	default:
		return zc.CompressionMethodStore
	}
}

func HeaderConvert(z *zip.FileHeader, b *zc.BasicZipFileHeader) {
	b.Name = z.Name
	b.Comment = z.Comment
	b.CompressionMethod = MethodConvert(z.Method)
	b.Modified = z.Modified
	b.Crc32 = z.CRC32
	b.OriginalSize = z.UncompressedSize64
	b.CompressedSize = z.CompressedSize64
}

func ReaderToBasicFile(
	r io.Reader,
	size uint64,
	maxBytes int64,
	b *zc.BasicZipFile,
) error {
	var maxSize int = min(int(size), int(maxBytes))
	var oldCap int = cap(b.Content)
	if oldCap < maxSize {
		b.Content = make([]byte, maxSize)
	}

	b.Content = b.Content[:maxSize]

	_, e := io.ReadFull(r, b.Content)
	return e
}

func ZipItemToBasic(
	z *zip.File,
	maxBytes int64,
	b *zc.BasicZipFile,
) error {
	HeaderConvert(&z.FileHeader, &b.Header)

	file, e := z.Open()
	if nil != e {
		return e
	}
	defer file.Close()

	limited := &io.LimitedReader{
		R: file,
		N: maxBytes,
	}

	return ReaderToBasicFile(limited, z.UncompressedSize64, maxBytes, b)
}

func ZipToBasicToOut(
	ctx context.Context,
	z *zip.Reader,
	maxBytes int64,
	b *zc.BasicZipFile,
	out bo.OutputBasic,
) error {
	for _, item := range z.File {
		e := ZipItemToBasic(item, maxBytes, b)
		if nil != e {
			return e
		}

		_, e = out(b)(ctx)
		if nil != e {
			return e
		}
	}
	return nil
}

func ZipFileToBasicToOut(
	ctx context.Context,
	f io.ReaderAt,
	fileSize int64,
	maxBytes int64,
	b *zc.BasicZipFile,
	out bo.OutputBasic,
) error {
	rdr, e := zip.NewReader(f, fileSize)
	if nil != e {
		return e
	}
	return ZipToBasicToOut(
		ctx,
		rdr,
		maxBytes,
		b,
		out,
	)
}

func ZipFilenameToBasicToOut(
	ctx context.Context,
	filename string,
	maxBytes int64,
	b *zc.BasicZipFile,
	out bo.OutputBasic,
) error {
	file, e := os.Open(filename)
	if nil != e {
		return e
	}
	defer file.Close()

	stat, e := file.Stat()
	if nil != e {
		return e
	}
	var size int64 = stat.Size()

	return ZipFileToBasicToOut(
		ctx,
		file,
		size,
		maxBytes,
		b,
		out,
	)
}

func ZipFilenamesToBasicToOut(
	ctx context.Context,
	filenames iter.Seq[string],
	maxBytes int64,
	out bo.OutputBasic,
) error {
	var buf zc.BasicZipFile
	for name := range filenames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		e := ZipFilenameToBasicToOut(
			ctx,
			name,
			maxBytes,
			&buf,
			out,
		)
		if nil != e {
			return e
		}
	}
	return nil
}

func MaxBytesToFilenamesToBasics(
	maxBytes int64,
) func(bo.OutputBasic) func(iter.Seq[string]) util.Io[util.Void] {
	return func(o bo.OutputBasic) func(iter.Seq[string]) util.Io[util.Void] {
		return func(names iter.Seq[string]) util.Io[util.Void] {
			return func(ctx context.Context) (util.Void, error) {
				return util.Empty, ZipFilenamesToBasicToOut(
					ctx,
					names,
					maxBytes,
					o,
				)
			}
		}
	}
}

func OutputBasicToFilenamesToBasicsDefault(
	o bo.OutputBasic,
) func(zipNames iter.Seq[string]) util.Io[util.Void] {
	return MaxBytesToFilenamesToBasics(MaxZipItemBytesDefault)(o)
}
