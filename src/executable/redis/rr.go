package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/go-redis/redis"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func choose(cmd *terminalw.Parser) interface{} {
	if cmd == nil {
		return nil
	}
	if cmd.ContainsFlagStrict("keys") {
		return showKeysAction(showKeys)
	}
	if cmd.ContainsFlagStrict("get") {
		return getKeyAction(getByKey)
	}
	if cmd.ContainsFlagStrict("hkeys") {
		return showHKeysAction(showHKeys)
	}
	if cmd.ContainsFlagStrict("hget") {
		return hGetByKeyAction(hGetByKey)
	}
	return nil
}

func showKeys(rdb *redis.Client, pattern string, count int64) []string {
	var res []string
	var err error
	var curRes []string
	curRes, _, err = rdb.Scan(0, pattern, count).Result()
	if err != nil {
		log.Fatalln(err)
	}
	res = append(res, curRes...)
	return res
}

func showHKeys(rdb *redis.Client, key string, count int64) []string {
	var res []string
	var err error
	var curRes []string
	curRes, _, err = rdb.HScan(key, 0, "*", count).Result()
	if err != nil {
		log.Fatalln(err)
	}
	res = append(res, curRes...)
	return res
}

func getByKey(rdb *redis.Client, key string) string {
	val, err := rdb.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return "nil"
		}
		panic(err)
	}
	return val
}

func hGetByKey(rdb *redis.Client, key string, fields ...string) []interface{} {
	if len(fields) == 0 {
		kvs := showHKeys(rdb, key, 100)
		for i, k := range kvs {
			if i%2 == 0 {
				fields = append(fields, k)
			}
		}
	}
	res, err := rdb.HMGet(key, fields...).Result()
	if err != nil {
		log.Fatalln(err)
	}
	var kvRes []interface{}
	for i, v := range res {
		kvRes = append(kvRes, fmt.Sprintf("%s:%s", fields[i], v))
	}
	return kvRes
}

type showKeysAction func(*redis.Client, string, int64) []string
type showHKeysAction func(*redis.Client, string, int64) []string
type getKeyAction func(*redis.Client, string) string
type hGetByKeyAction func(*redis.Client, string, ...string) []interface{}

func main() {
	var n = 10
	parser := terminalw.NewParser()
	parser.Bool("h", false, "print help info")
	parser.String("get", "", "get action")
	parser.String("hget", "", "hget action")
	parser.String("hkeys", "", "hkeys action")
	parser.Int("n", n, "max number of keys")
	parser.String("keys", "*", "show keys")
	parser.ParseArgsCmd("h")
	if parser.Empty() || parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	// Create a Redis client
	m := utilsw.GetAllConfig()
	if m == nil {
		fmt.Println("set redis.host and redis.password in ~/.configW")
		return
	}
	addr := m.GetOrDefault("redis.address", nil)
	password := m.GetOrDefault("redis.password", nil)
	if addr == nil {
		fmt.Println("set redis.address in ~/.configW")
		return
	}
	if password == nil {
		fmt.Println("set redis.password in ~/.configW")
		return
	}
	addrStr, passwordStr := addr.(string), password.(string)
	fmt.Printf("connected to %s\n~~~~~~~~~\n", color.GreenString(addrStr))

	rdb := redis.NewClient(&redis.Options{
		Addr:     addrStr,     // Default Redis port is 6379
		Password: passwordStr, // No password set
		DB:       0,           // Use default DB
	})

	act := choose(parser)
	switch act.(type) {
	// show keys
	case showKeysAction:
		pattern := parser.GetFlagValueDefault("keys", "*")
		if pattern == "" {
			pattern = "*"
		}
		keys := showKeys(rdb, pattern, int64(parser.GetIntFlagValOrDefault("n", n)))
		fmt.Printf("pattern: %s (%d found)\n", color.RedString(pattern), len(keys))
		for _, key := range keys {
			fmt.Printf("\t%s\n", key)
		}
	// show hkeys
	case showHKeysAction:
		key := parser.GetFlagValueDefault("hkeys", "")
		if key == "" {
			fmt.Println("field is required")
			return
		}
		fields := showHKeys(rdb, key, int64(parser.GetIntFlagValOrDefault("n", n)))
		fmt.Printf("hkeys: %s (%d found)\n", color.RedString(key), len(fields))
		var s string
		for i, key := range fields {
			if i%2 == 0 {
				s = key
			} else {
				s += ":" + key
				fmt.Printf("\t%s\n", s)
				s = ""
			}
		}
	// get
	case getKeyAction:
		key := parser.GetFlagValueDefault("get", "")
		val := getByKey(rdb, key)
		fmt.Printf("%s: %s\n", key, color.RedString(val))
	// hget
	case hGetByKeyAction:
		key := parser.GetFlagValueDefault("hget", "")
		if key == "" {
			fmt.Println("key is required")
			return
		}
		fields := parser.Positional.ToStringSlice()
		kvList := hGetByKey(rdb, key, fields...)
		fmt.Printf("hget key: %s\n", color.RedString(key))
		for _, kv := range kvList {
			fmt.Printf("\t%v\n", kv)
		}
	// others
	default:
		fmt.Println("not supported yet")
	}
}
