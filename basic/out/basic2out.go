package basic2out

import (
	zc "github.com/takanoriyanagitani/go-zip2cbor"
	util "github.com/takanoriyanagitani/go-zip2cbor/util"
)

type OutputBasic func(*zc.BasicZipFile) util.Io[util.Void]
