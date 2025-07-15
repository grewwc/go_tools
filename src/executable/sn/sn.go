package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	_helper "github.com/grewwc/go_tools/src/executable/sn/internal"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	bucketName = "the-best-bucket"
)

var (
	ak       string
	sk       string
	endpoint string
	bucket   *oss.Bucket
)

var (
	client *oss.Client
	err    error
)

func init() {
	config := utilsw.GetAllConfig()
	endpoint = config.GetOrDefault("oss.endpoint", "oss-cn-hangzhou.aliyuncs.com").(string)
	ak = config.GetOrDefault("oss.ak", "").(string)
	sk = config.GetOrDefault("oss.sk", "").(string)
	if ak == "" || sk == "" {
		panic("oss.ak/oss.sk is not set")
	}

	// init the oss client
	if client, err = oss.New(endpoint, ak, sk); err != nil {
		panic(err)
	}

	if bucket, err = client.Bucket(bucketName); err != nil {
		panic(err)
	}
}

func uploadSingleFile(wg *sync.WaitGroup, filename, ossKey string, force bool, retryCount int) {
	defer wg.Done()
	key := _helper.GetOssKey(ossKey)
	if key[len(key)-1] != '/' {
		if !force {
			if utilsw.PromptYesOrNo(fmt.Sprintf("do you want to overwrite the file: %s", key)) {
				fmt.Println(color.RedString("the file: %s will be overwritten!!", key))
			} else {
				fmt.Println("quit")
				return
			}
		}
	} else {
		key = strings.TrimSuffix(key, "/")
		key += "/" + filepath.Base(filename)
	}
	key = strings.TrimSuffix(key, "/")
	fmt.Printf(">>> uploading %s to %s\n", color.GreenString(filepath.Base(filename)), color.GreenString(key))
	if err = bucket.PutObjectFromFile(key, filename); err != nil {
		if retryCount <= 0 {
			panic(err)
		}
		retryCount--
		fmt.Println("retry...")
		wg.Add(1)
		uploadSingleFile(wg, filename, ossKey, force, retryCount)
	}
	fmt.Printf("<<< done uploading %s\n", color.GreenString(filepath.Base(filename)))
}

func upload(wg *sync.WaitGroup, filename, ossKey string, recursive, force bool) {
	if filename, err = filepath.Abs(filename); err != nil {
		panic(err)
	}
	if !utilsw.IsDir(filename) {
		wg.Add(1)
		go uploadSingleFile(wg, filename, ossKey, force, 3)
		return
	}

	filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if path == filename {
			return nil
		}
		subKey := strings.TrimSuffix(ossKey, "/") + "/" + strw.StripPrefix(strw.StripPrefix(path, filename), "/")
		upload(wg, path, subKey, recursive, force)
		return nil
	})

	fmt.Println("done")
}

func download(filename, ossKey string, ch chan struct{}, retryCount int) {
	defer func() {
		<-ch
	}()

	if filename, err = filepath.Abs(filename); err != nil {
		panic(err)
	}
	key := _helper.GetOssKey(ossKey)

	if utilsw.IsDir(filename) {
		filename += "/" + filepath.Base(key)
	} else if utilsw.IsExist(filename) { // 文件
		if utilsw.PromptYesOrNo(fmt.Sprintf("do you want to overwrite the file: %s", filename)) {
			fmt.Println(color.RedString("the file: %s will be overwritten!!", filename))
		} else {
			fmt.Println("quit")
			return
		}
	}
	fmt.Printf(">>> begin downloading %s \n", color.GreenString(filepath.Base(filename)))
	if err = bucket.GetObjectToFile(key, filename); err != nil {
		if retryCount <= 0 {
			panic(err)
		} else {
			log.Println(err)
			retryCount--
			ch <- struct{}{}
			download(filename, ossKey, ch, retryCount)
			return
		}
	}
	fmt.Printf("<<< done downloading %s \n", color.GreenString(filepath.Base(filename)))
}

func handleLs(args []string) {
	if len(args) == 0 {
		ls("", 0, 3)
	} else {
		for _, arg := range args {
			ls(arg, 2, 3)
		}
	}
}

// func handleDelete(objectKey string) {
// 	exist, err := bucket.IsObjectExist(objectKey)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if exist {
// 		res := utilsw.PromptYesOrNo(fmt.Sprintf("Delete %s? (y/n)", objectKey))
// 		if res {
// 			if err = bucket.DeleteObject(objectKey); err != nil {
// 				fmt.Println("Failed to delete", objectKey)
// 			} else {
// 				fmt.Printf("Deleted %s", objectKey)
// 			}
// 		} else {
// 			fmt.Println("Abort deleting", objectKey)
// 		}
// 	}
// }

func ls(dir string, prefixSpace, retryCount int) {
	result, err := bucket.ListObjectsV2()
	if err != nil {
		retryCount--
		if retryCount <= 0 {
			panic(err)
		} else {
			ls(dir, prefixSpace, retryCount)
			return
		}
	}
	if len(dir) > 0 && dir[len(dir)-1] != '/' {
		dir += "/"
	}
	s := cw.NewOrderedSet()
	for _, obj := range result.Objects {
		if strings.HasPrefix(obj.Key, dir) {
			name := strings.Repeat(" ", prefixSpace) + strw.StripPrefix(obj.Key, dir)
			if name == strings.Repeat(" ", prefixSpace) {
				continue
			}
			idx := strings.Index(name, "/")
			if idx < 0 {
				s.Add(name)
			} else {
				s.Add(color.HiCyanString(name[:idx]))
			}
		}
	}
	for name := range s.Iter().Iterate() {
		fmt.Println(name)
	}
}

func printHelp() {
	fmt.Println("sn cp test1.pdf test2.pdf test3.jpg oss://key")
	fmt.Println("sn cp oss://key ./")
}

func handleDelete(args []string) {
	if len(args) != 1 {
		printHelp()
	}
	deleteSingle(args[0])
}

func deleteSingle(ossKey string) {
	key := _helper.GetOssKey(ossKey)
	fmt.Printf(">>> begin deleting %s\n", color.RedString(key))
	if err = bucket.DeleteObject(key); err != nil {
		fmt.Fprintf(os.Stderr, "<<< failed to delete %s\n", key)
		return
	}
	fmt.Printf("<<< done deleting %s\n", color.RedString(key))
}

func handleCp(args []string, recursive, force bool) {
	if len(args) < 2 {
		printHelp()
		return
	}
	n := len(args)
	firstName := args[0]
	lastName := args[n-1]
	// upload
	if strings.HasPrefix(lastName, "oss://") {
		wg := sync.WaitGroup{}
		for _, name := range args[:n-1] {
			upload(&wg, name, lastName, recursive, force)
		}
		wg.Wait()
	} else if strings.HasPrefix(firstName, "oss://") { // download
		// 10 parallism
		ch := make(chan struct{}, 10)
		for _, name := range args[1:n] {
			ch <- struct{}{}
			download(name, firstName, ch, 3)
		}
		close(ch)
	} else {
		printHelp()
	}
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("r", false, "recursive")
	parser.Bool("f", false, "force")
	parser.ParseArgsCmd("r", "f")
	if parser.Empty() {
		printHelp()
		return
	}

	args := parser.Positional.ToStringSlice()
	if len(args) < 1 {
		printHelp()
		return
	}
	cmd := args[0]
	switch cmd {
	case "cp":
		handleCp(args[1:], parser.ContainsFlag("r"), parser.ContainsFlag("f"))
	case "ls":
		handleLs(args[1:])
	case "rm":
		handleDelete(args[1:])
	}
}
