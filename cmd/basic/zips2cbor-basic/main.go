package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"iter"
	"log"
	"os"

	zc "github.com/takanoriyanagitani/go-zip2cbor"
	util "github.com/takanoriyanagitani/go-zip2cbor/util"

	nr "github.com/takanoriyanagitani/go-zip2cbor/zip/names/reader"

	bo "github.com/takanoriyanagitani/go-zip2cbor/basic/out"
	ca "github.com/takanoriyanagitani/go-zip2cbor/basic/out/cbor/amacker"
	bs "github.com/takanoriyanagitani/go-zip2cbor/basic/std"
)

type IoConfig struct {
	io.Reader
	io.Writer
}

func (i IoConfig) ToZipNames() iter.Seq[string] {
	return nr.ReaderToNames(i.Reader)
}

func (i IoConfig) ToBasicWriter() (
	writer func(*zc.BasicZipFile) util.Io[util.Void],
	finalize func() error,
) {
	var bw *bufio.Writer = bufio.NewWriter(i.Writer)

	writer = ca.WriterToOutput(bw)
	finalize = func() error {
		return bw.Flush()
	}
	return
}

func (i IoConfig) ToBasicWriterChan(
	ctx context.Context,
	finalized chan<- error,
) func(*zc.BasicZipFile) util.Io[util.Void] {
	wtr, finalizer := i.ToBasicWriter()

	go func() {
		defer close(finalized)
		<-ctx.Done()

		e := finalizer()
		finalized <- e
	}()

	return wtr
}

var ioConfig IoConfig = IoConfig{
	Reader: os.Stdin,
	Writer: os.Stdout,
}

func sub(ctx context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	finalized := make(chan error)
	var wtr func(*zc.BasicZipFile) util.Io[util.Void] = ioConfig.
		ToBasicWriterChan(ctx, finalized)

	var out bo.OutputBasic = bo.OutputBasic(wtr)
	var names2out func(iter.Seq[string]) util.Io[util.Void] = bs.
		OutputBasicToFilenamesToBasicsDefault(out)

	var names iter.Seq[string] = ioConfig.ToZipNames()
	_, e := names2out(names)(ctx)
	cancel()

	var fe error = <-finalized

	return errors.Join(e, fe)
}

func main() {
	e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
