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

	"github.com/coocood/freecache"
	"github.com/schollz/progressbar/v3"
	auth "golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var (
	listUrl, movieUrl string
	procnum           int64
	onAuth            bool
	accessToken       string
	cache             *freecache.Cache
)

func init() {

	cache = freecache.NewCache(1024 * 1024 * 10)

	flag.StringVar(&listUrl, "listurl", "", "")
	flag.StringVar(&movieUrl, "movieurl", "", "")
	flag.Int64Var(&procnum, "procnum", 2, "")
	flag.BoolVar(&onAuth, "auth", false, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ex: go run . -listurl=$LIST_URL -movieurl=https://example.com/user -procnum 10\n")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func getUrlList() []map[string]string {
	res, err := http.Get(listUrl)
	if err != nil {
		log.Fatal("Get error", err)
	}
	defer res.Body.Close()

	m := []map[string]string{}
	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		log.Fatal("JSON encode error", err)
	}
	return m
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	e, ctx := errgroup.WithContext(ctx)

	list := getUrlList()

	sem := semaphore.NewWeighted(procnum)

	accessToken := getAccessToken()
	for _, l := range list {
		url := l["dst"]
		sem.Acquire(ctx, 1)
		e.Go(func() error {
			doSomething(ctx, url, accessToken, onAuth)
			sem.Release(1)
			return nil
		})
	}

	if err := e.Wait(); err != nil {
		log.Println(err)
	}
}

func doSomething(ctx context.Context, url string, accessToken string, onAuth bool) {

	time.Sleep(time.Second * 1)
	fullUrl := fmt.Sprintf("%s/%s", movieUrl, url)

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		log.Print(err)
		return
	}
	if onAuth {
		token := getAccessToken()
		fmt.Println("token", token)
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	} else {
		defer res.Body.Close()
		f := io.Discard
		ch := make(chan struct{})
		bar := getBar(ch, res.ContentLength)
		_, err = io.Copy(io.MultiWriter(f, bar), res.Body)
		if err != nil {
			log.Println(err)
		}
		if res.StatusCode != 200 {
			log.Println("Error", res.StatusCode, req.Header, fullUrl)
		}
	}

	ctx.Done()
}

func getBar(ch chan struct{}, contentLength int64) *progressbar.ProgressBar {
	bar := progressbar.DefaultBytes(
		contentLength,
		"downloading",
	)
	return bar
}

func getAccessToken() string {

	if accessToken != "" {
		return accessToken
	}

	ctx := context.Background()
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
	credentials, err := auth.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		log.Println("Credential Error", err)
		return ""
	}
	t, _ := credentials.TokenSource.Token()
	if err != nil {
		fmt.Print(err)
	}
	accessToken = t.AccessToken

	return accessToken

}
