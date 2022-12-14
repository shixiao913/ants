package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
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

type Parser interface {
	Parse(body []byte) ([]string, error)
}

var headerRe = regexp.MustCompile(`{"contId"[\s\S]*?,"name":"([\s\S]*?)",`)

type ReParser struct {
}

func (p *ReParser) Parse(body []byte) ([]string, error) {
	matches := headerRe.FindAllSubmatch(body, -1)

	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, string(match[1]))
	}

	return result, nil
}

type XPathParser struct {
}

func (p *XPathParser) Parse(body []byte) ([]string, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	nodes := htmlquery.Find(doc, `//div[@class="index_carousel_img__HbOWM"]/a[@target="_blank"]/img`)

	result := make([]string, 0, len(nodes))

	for _, node := range nodes {
		msg := htmlquery.SelectAttr(node, "alt")
		result = append(result, msg)
	}

	return result, nil
}

/*
CS stands for CSS Selector
*/
type CSParser struct {
}

func (p *CSParser) Parse(body []byte) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	s := doc.Find("div.index_carousel_img__HbOWM a[target=_blank] img")

	result := make([]string, 0, len(s.Nodes))
	for _, node := range s.Nodes {
		msg := htmlquery.SelectAttr(node, "alt")
		result = append(result, msg)
	}

	return result, nil
}

func main() {
	url := "https://www.thepaper.cn"

	body, err := fetch(url)
	if err != nil {
		fmt.Println("fetch error: ", err)
		return
	}

	//parser := &ReParser{}
	//parser := &XPathParser{}
	parser := &CSParser{}

	news, err := parser.Parse(body)
	if err != nil {
		fmt.Println("parser error: ", err)
		return
	}

	for _, msg := range news {
		fmt.Println("fetched news: ", msg)
	}
}
