package zip2cbor

import (
	"time"
)

type CompressionMethod uint16

const (
	CompressionMethodStore   CompressionMethod = 0
	CompressionMethodDeflate CompressionMethod = 8
)

type BasicZipFileHeader struct {
	// Modified time of the file.
	Modified time.Time `json:"modified"`

	// The filepath of the file in the zip file.
	Name string `json:"name"`

	// A user defined comment.
	Comment string `json:"comment"`

	// The compressed size of the file in bytes.
	CompressedSize uint64 `json:"compressed_size"`

	// The original size of the file in bytes.
	OriginalSize uint64 `json:"original_size"`

	// The checksum of the file.
	Crc32 uint32 `json:"crc32"`

	CompressionMethod `json:"method"`
}

type BasicZipFile struct {
	Header  BasicZipFileHeader `json:"header"`
	Content []byte             `json:"content"`
}
