package main

import (
	"bytes"
	"log"
	"os"
	"sort"
)

func main() {
	{
		data, err := os.ReadFile("./with_context.go")
		if err != nil {
			log.Fatal(err)
		}

		replacements := map[string]string{
			"WithContext[T":       "[T", // Func Signature
			"WithContext ":       " ",   // Comment
			"ctx context.Context": "done <-chan struct{}",
			"WithContext(ctx":     "(done",
			"context canceled":    "done chan closed",
			"ctx.Done()":          "done",
			"\"context\"\n\t":     "", // meybe better to use goimports?
			`//go:generate go run ./cmd/internal/gendone/`: "",
			`ctx == nil`:            `done == nil`,
			`ErrContext`:            `ErrDone`,
			`nil Contex`:            `nil done chan`,
			`or context cancelled.`: "or done is closed",
		}

		for _, search := range sortKeyDesc(replacements) {
			data = bytes.ReplaceAll(data, []byte(search), []byte(replacements[search]))
		}
		// docs related
		data = bytes.ReplaceAll(data, []byte("WithContext "), []byte("WithDone "))

		data = append([]byte("// Code generated by cmd/internal/gendone. DO NOT EDIT.\n"), data...)
		os.WriteFile("./with_done.go", data, 0x644)
	}

	// ---
	{
		data, err := os.ReadFile("./with_context_test.go")
		if err != nil {
			log.Fatal(err)
		}

		replacements := map[string]string{

			"WithContext(":                 "(",
			"WithContextNil(":              "_WhereDoneNil(",
			"testTableContext":             "testTableDone",
			"ctx, cancel := test.fncCtx()": "done, cancel := test.fncDone()",
			"<-ctx.Done()":                 "<-done",
			"ctx context.Context":          "done chan struct{}",
			"ctx,":                         "done,",
			"\"context\"\n\t":              "", // meybe better to use goimports?
			`//go:generate go run ./cmd/internal/gendone/`: "",
			`ErrContext`: `ErrDone`,
		}
		for _, search := range sortKeyDesc(replacements) {
			data = bytes.ReplaceAll(data, []byte(search), []byte(replacements[search]))
		}

		data = append([]byte("// Code generated by cmd/internal/gendone. DO NOT EDIT.\n"), data...)
		os.WriteFile("./with_done_test.go", data, 0x644)

	}
}

func sortKeyDesc(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	sort.Slice(out, func(i, j int) bool {
		return len(out[j]) < len(out[i])
	})

	return out
}
