package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	_helpers "github.com/grewwc/go_tools/src/executable/ss/_helpers"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
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
	config := utilsW.GetAllConfig()
	endpoint = config.GetOrDefault("oss.endpoing", "oss-cn-hangzhou.aliyuncs.com").(string)
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

func uploadSingleFile(wg *sync.WaitGroup, filename, ossKey string, force bool) {
	defer wg.Done()
	key := _helpers.GetOssKey(ossKey)
	if key[len(key)-1] != '/' {
		if filepath.Base(key) != filepath.Base(filename) {
			key += "/" + filepath.Base(filename)
		} else if !force {
			if utilsW.PromptYesOrNo(fmt.Sprintf("do you want to overwrite the file: %s", key)) {
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
		panic(err)
	}
	fmt.Printf("<<< done uploading %s\n", color.GreenString(filepath.Base(filename)))
}

func upload(wg *sync.WaitGroup, filename, ossKey string, recursive, force bool) {
	if filename, err = filepath.Abs(filename); err != nil {
		panic(err)
	}
	if !utilsW.IsDir(filename) {
		wg.Add(1)
		go uploadSingleFile(wg, filename, ossKey, force)
		return
	}

	filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if path == filename {
			return nil
		}
		subKey := strings.TrimSuffix(ossKey, "/") + "/" + stringsW.StripPrefix(stringsW.StripPrefix(path, filename), "/")
		upload(wg, path, subKey, recursive, force)
		return nil
	})

	fmt.Println("done")
}

func download(filename, ossKey string, recursive, force bool, ch <-chan struct{}) {
	defer func() {
		<-ch
	}()

	if filename, err = filepath.Abs(filename); err != nil {
		panic(err)
	}
	key := _helpers.GetOssKey(ossKey)

	if utilsW.IsDir(filename) {
		filename += "/" + filepath.Base(key)
	} else if utilsW.IsExist(filename) { // 文件
		if utilsW.PromptYesOrNo(fmt.Sprintf("do you want to overwrite the file: %s", filename)) {
			fmt.Println(color.RedString("the file: %s will be overwritten!!", filename))
		} else {
			fmt.Println("quit")
			return
		}
	}
	fmt.Printf(">>> begin downloading %s \n", color.GreenString(filepath.Base(filename)))
	if err = bucket.GetObjectToFile(key, filename); err != nil {
		panic(err)
	}
	fmt.Printf("<<< done downloading %s \n", color.GreenString(filepath.Base(filename)))
}

func handleLs(args []string) {
	if len(args) == 0 {
		ls("", 0)
	} else {
		for _, arg := range args {
			ls(arg, 2)
		}
	}
}

func ls(dir string, prefixSpace int) {
	result, err := bucket.ListObjectsV2()
	if err != nil {
		panic(err)
	}
	if len(dir) > 0 && dir[len(dir)-1] != '/' {
		dir += "/"
	}
	s := containerW.NewOrderedSet()
	for _, obj := range result.Objects {
		if strings.HasPrefix(obj.Key, dir) {
			name := strings.Repeat(" ", prefixSpace) + stringsW.StripPrefix(obj.Key, dir)
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
	for name := range s.Iterate() {
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
	key := _helpers.GetOssKey(ossKey)
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
			download(name, firstName, recursive, force, ch)
		}
		close(ch)
	} else {
		printHelp()
	}
}

func main() {
	fs := flag.NewFlagSet("flag", flag.ExitOnError)
	fs.Bool("r", false, "recursive")
	fs.Bool("f", false, "force")

	parsed := terminalW.ParseArgsCmd("r", "f")
	if parsed == nil {
		printHelp()
		return
	}

	args := parsed.Positional.ToStringSlice()
	if len(args) < 1 {
		printHelp()
		return
	}
	cmd := args[0]
	switch cmd {
	case "cp":
		handleCp(args[1:], parsed.ContainsFlag("r"), parsed.ContainsFlag("f"))
	case "ls":
		handleLs(args[1:])
	case "rm":
		handleDelete(args[1:])
	}
}
