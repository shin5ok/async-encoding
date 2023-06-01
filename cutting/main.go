package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/mowshon/moviego"
)

var (
	projectID       = os.Getenv("GOOGLE_CLOUD_PROJECT")
	subscName       = os.Getenv("SUBSCRIPTION")
	bucketName      = os.Getenv("BUCKET")
	firestoreClient *firestore.Client
	myIpAddr        []byte
)

const findIPUrl = `http://ifconfig.me`

func init() {
	ctx := context.Background()
	firestoreClient, _ = firestore.NewClient(ctx, projectID)

	httpClient := http.Client{Timeout: 2 * time.Second}
	resp, err := httpClient.Get(findIPUrl)
	if err != nil {
		log.Println("cannot get to external site")
	}
	defer resp.Body.Close()

	myIpAddr, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println("cannot get IP")
	}
}

func main() {
	err := pullAndConvert(projectID, subscName)
	defer firestoreClient.Close()
	if err != nil {
		log.Panicln(err)
	}
}

func pullAndConvert(projectID, subscName string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subscName)
	sub.ReceiveSettings.Synchronous = false
	sub.ReceiveSettings.NumGoroutines = 2
	sub.ReceiveSettings.MaxOutstandingMessages = 1

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, doConvert)
	if err != nil {
		return fmt.Errorf("sub.Receive returned error: %v", err)
	}

	return nil

}

type params struct {
	Src         string  `json:"src"`
	Dst         string  `json:"dst"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	UserID      string  `json:"user_id"`
	ProcessHost string
}

func doConvert(ctx context.Context, msg *pubsub.Message) {

	log.Printf("Start processing id %s\n", msg.ID)

	var data params
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Println("json decode error", err)
		msg.Ack() // decode error would be critical, so should be given up
		return
	}

	src := data.Src
	dst := msg.ID + ".mp4"
	srcTmp := msg.ID + "-" + src
	dstTmp := msg.ID + "-" + dst

	err := downloadFile(bucketName, src, srcTmp)
	if err != nil {
		log.Println(err)
		msg.Nack()
		return
	}

	defer func() {
		os.Remove(srcTmp)
		os.Remove(dstTmp)
		log.Println("cleanup", srcTmp, dstTmp)
	}()

	log.Printf("%+v\n", data)
	first, err := moviego.Load(srcTmp)

	if err != nil {
		log.Println(err)
		msg.Nack()
		return
	}

	err = first.SubClip(data.Start, data.End).Output(dstTmp).Run()
	if err != nil {
		log.Println(err)
		msg.Ack()
		return
	}

	dstFull := fmt.Sprintf("%s/%s.mp4", data.UserID, msg.ID)
	uploadFile(bucketName, dstTmp, dstFull)

	outFile := fmt.Sprintf("gs://%s/%s", bucketName, dstFull)
	log.Println("out", outFile)

	data.ProcessHost = string(myIpAddr)
	data.Dst = dstFull

	err = register2DB(data)
	if err != nil {
		log.Println(err)
		msg.Nack()
		return
	}
	log.Println("entry has been registered to DB")

	msg.Ack()
}

func register2DB(data params) error {

	ctx := context.Background()
	doc := data
	firestoreClient.Collection("data").Doc(data.UserID).Set(ctx, doc)
	return nil
}

func downloadFile(bucket, object string, dst string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("f.Close: %v", err)
	}

	return nil

}

func uploadFile(bucket, src, object string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	o = o.If(storage.Conditions{DoesNotExist: true})
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}
