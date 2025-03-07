package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
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
		buf = stringsW.StringToBytes(utilsW.ReadClipboardText())
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
		imgBytes = stringsW.StringToBytes(utilsW.ReadClipboardText())
	} else {
		imgBytes, err = os.ReadFile(fname)
		if err != nil {
			panic(err)
		}
	}

	s := ";base64,"
	indices := stringsW.KmpSearchBytes(imgBytes, stringsW.StringToBytes(s))
	if len(indices) == 1 {
		imgBytes = imgBytes[indices[0]+len(s):]
	}
	img := stringsW.BytesToString(imgBytes)
	b, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(outName, b, 0666); err != nil {
		panic(err)
	}
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("f", true, "get content from file")
	fs.Bool("c", false, "get content from clipboard")
	fs.String("out", "", "output file name")
	fs.Bool("toimg", false, "")

	parsed := terminalW.ParseArgsCmd("f", "toimg", "c")
	if parsed.ContainsAnyFlagStrict("h") {
		fs.PrintDefaults()
		return
	}
	isURL := true
	toImage := parsed.ContainsFlagStrict("toimg")
	if parsed.ContainsAnyFlagStrict("f", "c") {
		isURL = false
	}
	if parsed.GetFlagValueDefault("out", "") != "" {
		outName = parsed.GetFlagValueDefault("out", "")
	}
	pos := parsed.Positional.ToStringSlice()
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
