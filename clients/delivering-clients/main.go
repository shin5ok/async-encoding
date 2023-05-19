package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var (
	listUrl, movieUrl string
	num               int64
)

func init() {

	flag.StringVar(&listUrl, "listurl", "", "")
	flag.StringVar(&movieUrl, "movieurl", "", "")
	flag.Int64Var(&num, "num", 2, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ex: go run . -listurl=$LIST_URL -movieurl=https://example.com -num 10\n")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func getUrlList() []map[string]string {
	res, err := http.Get(listUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	m := []map[string]string{}
	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	e, ctx := errgroup.WithContext(ctx)

	list := getUrlList()

	sem := semaphore.NewWeighted(num)

	for _, l := range list {
		url := l["dst"]
		sem.Acquire(ctx, 1)
		e.Go(func() error {
			doSomething(ctx, url)
			sem.Release(1)
			return nil
		})
	}

	if err := e.Wait(); err != nil {
		log.Println(err)
	}
}

func doSomething(ctx context.Context, url string) {

	time.Sleep(time.Second * 1)
	fullUrl := fmt.Sprintf("%s/%s", movieUrl, url)

	fmt.Println(fullUrl)

	res, err := http.Get(fullUrl)
	if err != nil {
		log.Println(err)
	} else {
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}
	}

	ctx.Done()
}
