package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/grewwc/go_tools/src/terminalW"
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
		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
	} else {
		f, err := os.Open(url)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		buf, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}
	str := base64.StdEncoding.EncodeToString(buf)
	if err = ioutil.WriteFile(outName, []byte(str), 0666); err != nil {
		panic(err)
	}
}

func base64ToImage(fname, outName string){
	imgBytes, err := ioutil.ReadFile(fname)
	if err != nil{
		panic(err)
	}
	img := string(imgBytes)
	b, err := base64.StdEncoding.DecodeString(img)
	if err != nil{
		panic(err)
	}

	if err = ioutil.WriteFile(outName, b, 0666); err != nil{
		panic(err)
	}
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("f", true, "pass file ")
	fs.String("out", "", "output file name")
	fs.Bool("toimg", false, "")

	parsed := terminalW.ParseArgsCmd("f", "toimg")
	if parsed == nil || parsed.ContainsAnyFlagStrict("h") {
		fs.PrintDefaults()
		return
	}
	isURL := true
	if parsed.ContainsFlagStrict("f") {
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
	if parsed.ContainsFlagStrict("toimg"){
		if isURL{
			panic("must pass file name")
		}
		if outName == "image-base64.txt"{
			outName = "output.jpg"
		}
		base64ToImage(pos[0], outName)
		return
	}
	transferImgToBase64(pos[0], isURL)
}
