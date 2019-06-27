package main

import (
	"context"
	"strings"
)

func toUpper(ctx context.Context, input string) (output string, err error) {
	output = strings.ToUpper(input)
	return
}
