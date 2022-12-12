package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	r := bufio.NewReader(resp.Body)
	e := determineCharSet(r)
	utf8Reader := transform.NewReader(r, e.NewDecoder())
	return io.ReadAll(utf8Reader)
}

func determineCharSet(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		fmt.Println("fetch error:", err)
		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}

var headerRe = regexp.MustCompile(`{"contId"[\s\S]*?,"name":"([\s\S]*?)",`)

func main() {
	url := "https://www.thepaper.cn"

	body, err := fetch(url)
	if err != nil {
		fmt.Println("fetch error: ", err)
		return
	}

	matches := headerRe.FindAllSubmatch(body, -1)

	for _, m := range matches {
		fmt.Println("fetched news: ", string(m[1]))
	}
}
