package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/schollz/progressbar/v3"
)

var (
	postUrl         string
	procNum         int64
	requestNum      int
	list            []string
	defaultListFile = "../movies.txt"
	isDebug         = os.Getenv("DEBUG")
	listFile        string
)

func debugPrint(message any) {
	if isDebug != "" {
		log.Println(message)
	}
}

func init() {

	flag.StringVar(&postUrl, "posturl", "", "")
	flag.StringVar(&listFile, "listfile", defaultListFile, "")
	flag.Int64Var(&procNum, "procnum", 10, "")
	flag.IntVar(&requestNum, "requestnum", 100, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ex: go run . -posturl=https://example.com/request -listfile ./movies.txt -procnum 100 -requestnum 1000\n")
	}

	flag.Parse()
	list = func() []string {
		f, err := os.Open(listFile)
		if err != nil {
			log.Fatal(err)
		}

		var data []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			text := scanner.Text()
			data = append(data, text)
		}
		return data

	}()
}

func genParams() map[string]any {
	start := rand.Intn(30)
	end := start + 5
	index := rand.Intn(len(list))
	src := list[index]
	dst := ""
	id, _ := uuid.NewRandom()
	params := map[string]any{
		"src":     src,
		"dst":     dst,
		"start":   start,
		"end":     end,
		"user_id": id.String(),
	}
	return params
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	e, ctx := errgroup.WithContext(ctx)

	sem := semaphore.NewWeighted(procNum)

	bar := progressbar.Default(int64(requestNum))

	for i := 0; i < requestNum; i++ {
		bar.Add(1)
		sem.Acquire(ctx, 1)
		params := genParams()
		debugPrint(params)
		e.Go(func() error {
			doSomething(ctx, postUrl, params)
			sem.Release(1)
			return nil
		})
	}

	if err := e.Wait(); err != nil {
		log.Println(err)
	}

}

func doSomething(ctx context.Context, url string, data map[string]any) {

	params := data
	dataJson, _ := json.Marshal(params)

	debugPrint(string(dataJson))
	res, err := http.Post(url, "application/json", bytes.NewReader(dataJson))

	if err != nil {
		log.Println(err)
	} else {
		defer res.Body.Close()
		dataBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}
		_ = dataBody
		debugPrint(string(dataBody))
	}

	ctx.Done()
}
