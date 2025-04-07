package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grewwc/go_tools/src/strW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/typesW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var (
	outName = "image-base64.txt"
)

func transferImgToBase64(url string, isUrl bool) {
	var err error
	var buf []byte
	if isUrl {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		buf, err = io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
	} else if url != "" {
		f, err := os.Open(url)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		buf, err = io.ReadAll(f)
		if err != nil {
			panic(err)
		}
	} else { // clipboard
		buf = typesW.StrToBytes(utilsW.ReadClipboardText())
	}
	str := base64.StdEncoding.EncodeToString(buf)
	if err = os.WriteFile(outName, []byte(str), 0666); err != nil {
		panic(err)
	}
}

func base64ToImage(fname, outName string) {
	var imgBytes []byte
	var err error
	if fname == "" {
		imgBytes = typesW.StrToBytes(utilsW.ReadClipboardText())
	} else {
		imgBytes, err = os.ReadFile(fname)
		if err != nil {
			panic(err)
		}
	}

	s := ";base64,"
	indices := strW.KmpSearchBytes(imgBytes, typesW.StrToBytes(s), 1)
	if len(indices) == 1 {
		imgBytes = imgBytes[indices[0]+len(s):]
	}
	img := typesW.BytesToStr(imgBytes)
	b, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(outName, b, 0666); err != nil {
		panic(err)
	}
}

func main() {
	parser := terminalW.NewParser()
	parser.Bool("f", true, "get content from file")
	parser.Bool("c", false, "get content from clipboard")
	parser.String("out", "", "output file name")
	parser.Bool("toimg", false, "")
	parser.Bool("h", false, "print help msg")
	parser.ParseArgsCmd("f", "toimg", "c", "h")
	if parser.ContainsAnyFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	isURL := true
	toImage := parser.ContainsFlagStrict("toimg")
	if parser.ContainsAnyFlagStrict("f", "c") {
		isURL = false
	}
	if parser.GetFlagValueDefault("out", "") != "" {
		outName = parser.GetFlagValueDefault("out", "")
	}
	pos := parser.Positional.ToStringSlice()
	if len(pos) > 1 {
		fmt.Println("only 1 positional arg allowed")
		return
	}
	if len(pos) == 0 {
		pos = []string{""}
		toImage = true
		isURL = false
	}
	if toImage {
		if isURL {
			panic("must pass file name")
		}
		if outName == "image-base64.txt" {
			outName = "output.jpg"
		}
		base64ToImage(pos[0], outName)
		return
	}
	transferImgToBase64(pos[0], isURL)
}
