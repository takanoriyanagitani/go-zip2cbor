package util

import (
	"context"
)

type Io[T any] func(context.Context) (T, error)

type Void struct{}

var Empty Void = struct{}{}
