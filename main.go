package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-colorable"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	GREY    = "\033[0;30;1m"
	RED     = "\033[0;31;1m"
	GREEN   = "\033[0;32;1m"
	YELLOW  = "\033[0;33;1m"
	BLUE    = "\033[0;34;1m"
	MAGENTA = "\033[0;35;1m"
	CYAN    = "\033[0;36;1m"
	WHITE   = "\033[0;37;1m"
	END     = "\033[0m"
)

func displayContent(w io.Writer, content *goquery.Selection) {
	if content.Is("ol") || content.Is("ul") {
		li := content.Find("li")
		if li.Size() > 0 {
			li.Each(func(i int, l *goquery.Selection) {
				fmt.Fprintln(w, GREY+strconv.Itoa(i+1)+"."+END+" "+l.Text())
			})
		} else {
			fmt.Fprintln(w, content.Text())
		}
	} else {
		fmt.Fprintln(w, content.Text())
	}
}

var opts struct {
	Word  string `positional-arg-name:"word"`
	Range string `long:"range" short:"r" description:"range of showing" default:"1:3"`
}

func main() {
	stdOut := bufio.NewWriter(colorable.NewColorableStdout())

	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "alc"
	parser.Usage = "WORD [OPTIONS]"

	args, err := parser.Parse()
	if err != nil {
		return
	}

	word := url.QueryEscape(strings.Join(args, " "))
	doc, err := goquery.NewDocument("https://eow.alc.co.jp/search?q=" + word)
	if err != nil {
		panic(err)
	}

	doc.Find(".ex_sentence, .kana").Each(func(i int, ex *goquery.Selection) {
		ex.Remove()
	})

	rxrange := regexp.MustCompile(`(\d+):(\d+)`)
	rxlvl := regexp.MustCompile(`【レベル】\d+、`)
	rxkana := regexp.MustCompile(`【＠】`)

	begin, err := strconv.Atoi(rxrange.ReplaceAllString(opts.Range, "$1"))
	if err != nil {
		panic(err)
	}
	end, err := strconv.Atoi(rxrange.ReplaceAllString(opts.Range, "$2"))
	if err != nil {
		panic(err)
	}
	doc.Find("#resultsList").Each(func(i int, result *goquery.Selection) {
		result.Find(".midashi, .midashi_je").Each(func(_i int, midashi *goquery.Selection) {
			if begin <= _i+1 && _i+1 <= end {
				fmt.Fprintln(stdOut, RED+midashi.Text()+END)
				content := midashi.Next()
				content.Find(".attr").Each(func(i int, attr *goquery.Selection) {
					str := rxlvl.ReplaceAllString(attr.Text(), "")
					str = rxkana.ReplaceAllString(str, "【カナ】")
					fmt.Fprintln(stdOut, YELLOW+str+END)
				})
				if content.Is("div") {
					wordclass := content.Find(".wordclass")
					if wordclass.Size() > 0 {
						wordclass.Each(func(__i int, w *goquery.Selection) {
							fmt.Fprintln(stdOut, BLUE+w.Text()+END)
							displayContent(stdOut, w.Next())
						})
					} else {
						list := content.Find("ul, ol")
						if list.Size() > 0 {
							displayContent(stdOut, list)
						} else {
							displayContent(stdOut, content)
						}
					}
				}
			}
		})
	})

	stdOut.Flush()
}
