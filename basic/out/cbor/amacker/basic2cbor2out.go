package cbor2out

import (
	"context"
	"io"

	ac "github.com/fxamacker/cbor/v2"

	zc "github.com/takanoriyanagitani/go-zip2cbor"
	util "github.com/takanoriyanagitani/go-zip2cbor/util"
)

func WriterToOutput(w io.Writer) func(*zc.BasicZipFile) util.Io[util.Void] {
	var enc *ac.Encoder = ac.NewEncoder(w)
	return func(b *zc.BasicZipFile) util.Io[util.Void] {
		return func(_ context.Context) (util.Void, error) {
			return util.Empty, enc.Encode(b)
		}
	}
}
