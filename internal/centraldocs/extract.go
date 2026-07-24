package centraldocs

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

const maxNextDataSize = 16 << 20

func extractNextData(r io.Reader) ([]byte, error) {
	tokenizer := html.NewTokenizer(r)
	var data []byte
	found := 0
	inside := false
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("parse Central page HTML: %w", err)
			}
			if found == 0 {
				return nil, errors.New("Central page is missing script#__NEXT_DATA__")
			}
			if found > 1 {
				return nil, errors.New("Central page contains duplicate script#__NEXT_DATA__ elements")
			}
			if len(bytes.TrimSpace(data)) == 0 {
				return nil, errors.New("Central page script#__NEXT_DATA__ is empty")
			}
			return data, nil
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data != "script" {
				continue
			}
			for _, attribute := range token.Attr {
				if attribute.Key == "id" && attribute.Val == "__NEXT_DATA__" {
					found++
					inside = true
					break
				}
			}
		case html.TextToken:
			if inside {
				chunk := tokenizer.Text()
				if len(data)+len(chunk) > maxNextDataSize {
					return nil, fmt.Errorf("Central page script#__NEXT_DATA__ exceeds %d bytes", maxNextDataSize)
				}
				data = append(data, chunk...)
			}
		case html.EndTagToken:
			if tokenizer.Token().Data == "script" {
				inside = false
			}
		}
	}
}
