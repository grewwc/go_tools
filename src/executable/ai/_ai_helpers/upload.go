package _ai_helpers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/sortW"
	"github.com/grewwc/go_tools/src/utilsW"
)

type comparedTuple containerW.Tuple

func (tup *comparedTuple) Compare(other interface{}) int {
	otherTup, ok := other.(*comparedTuple)
	if !ok {
		return 0
	}
	return ((*containerW.Tuple)(tup)).Get(0).(int) - ((*containerW.Tuple)(otherTup)).Get(0).(int)
}

func uploadSingleQwenlongFile(apiKey, filename string) (string, error) {
	defer func() {
		recover()
	}()
	client := &http.Client{}
	baseUrl := "https://dashscope.aliyuncs.com/compatible-mode/v1/files"
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)
	writer.WriteField("purpose", "file-extract")

	writer.Close()
	req, err := http.NewRequest("POST", baseUrl, &body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	fmt.Println("Uploading file...")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	j := utilsW.NewJsonFromByte(b)
	fileid := j.GetString("id")
	fmt.Println("Finished upload. Fileid: ", fileid)
	return fileid, nil
}

func UploadQwenLongFiles(apiKey string, files []string) []string {
	ch := make(chan *comparedTuple, len(files))
	defer close(ch)
	for i, file := range files {
		file = strings.TrimSpace(file)
		file = utilsW.ExpandUser(file)
		go upload(ch, apiKey, file, i)
	}
	resultTuple := make([]*comparedTuple, 0, len(files))
	for i := 0; i < len(files); i++ {
		resultTuple = append(resultTuple, <-ch)
	}
	sortW.QuickSortComparable[*comparedTuple](resultTuple)

	result := make([]string, len(resultTuple))
	for i, tup := range resultTuple {
		result[i] = ((*containerW.Tuple)(tup)).Get(1).(string)
	}
	return result

}

func upload(result chan<- *comparedTuple, apiKey, filename string, order int) {
	fileid := uploadSingleQwenLongFileWithRetry(apiKey, filename, 5)
	result <- (*comparedTuple)(containerW.NewTuple(order, fileid))
}

func uploadSingleQwenLongFileWithRetry(apiKey, filename string, retry int) string {
	for i := 0; i < retry; i++ {
		fileid, err := uploadSingleQwenlongFile(apiKey, filename)
		if err == nil && fileid != "" {
			return fileid
		}
	}
	return ""
}
