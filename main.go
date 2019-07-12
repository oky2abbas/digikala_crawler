package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	url := os.Args[1]
	if downloadFromURL(url) {
		parsHTML("f.html")
	}
}

func downloadFromURL(url string) bool {
	// tokens := strings.Split(url, "/")
	fileName := "f.html" //tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return false
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return false
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return false
	}

	fmt.Println(n, "bytes downloaded.")

	return true
}

type SkipTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewSkipTillReader(reader io.Reader, delim []byte) *SkipTillReader {
	return &SkipTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (str *SkipTillReader) Read(p []byte) (n int, err error) {
	if str.found {
		return str.rdr.Read(p)
	} else {
		// search byte by byte for the delimiter
	outer:
		for {
			for i := range str.delim {
				var c byte
				c, err = str.rdr.ReadByte()
				if err != nil {
					n = 0
					return
				}
				// doens't match so start over
				if str.delim[i] != c {
					continue outer
				}
			}
			str.found = true
			// we read the delimiter so add it back
			str.rdr = bufio.NewReader(io.MultiReader(bytes.NewReader(str.delim), str.rdr))
			return str.Read(p)
		}
	}
}

type ReadTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewReadTillReader(reader io.Reader, delim []byte) *ReadTillReader {
	return &ReadTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (rtr *ReadTillReader) Read(p []byte) (n int, err error) {
	if rtr.found {
		return 0, io.EOF
	} else {
	outer:
		for n < len(p) {
			for i := range rtr.delim {
				var c byte
				c, err = rtr.rdr.ReadByte()
				if err != nil && n > 0 {
					err = nil
					return
				} else if err != nil {
					return
				}
				p[n] = c
				n++
				if rtr.delim[i] != c {
					continue outer
				}
			}
			rtr.found = true
			break
		}
		if n == 0 {
			err = io.EOF
		}
		return
	}
}

func parsHTML(name string) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Print(err)
	}

	ss := string(b)
	r := strings.NewReader(ss)
	start := "<div class=\"c-product__seller-price-raw js-price-value\">"
	end := "</div>"
	str := NewSkipTillReader(r, []byte(start))
	rtr := NewReadTillReader(str, []byte("</span>"))
	bs, err := ioutil.ReadAll(rtr)

	sts := strings.Replace(string(bs), start, "", -1)
	sts = strings.Replace(sts, end, "", -1)

	fmt.Println(sts, err)
}
