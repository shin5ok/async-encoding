package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/mowshon/moviego"
)

var (
	projectID  = os.Getenv("GOOGLE_CLOUD_PROJECT")
	subscName  = os.Getenv("SUBSCRIPTION")
	bucketName = os.Getenv("BUCKET")
)

func main() {
	pullAndConvert(projectID, subscName)
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
	//sub.ReceiveSettings.NumGoroutines = 16
	//sub.ReceiveSettings.MaxOutstandingMessages = 8

	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, doConvert)
	if err != nil {
		return fmt.Errorf("sub.Receive returned error: %v", err)
	}

	return nil

}

type params struct {
	Src    string  `json:"src"`
	Dst    string  `json:"dst"`
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	UserID string  `json:"user_id"`
}

func doConvert(ctx context.Context, msg *pubsub.Message) {

	var data params
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Println(err)
		return
	}

	src := data.Src
	dst := data.Dst

	err := downloadFile(bucketName, src, src)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		os.Remove(src)
		os.Remove(dst)
		log.Println("cleanup", src, dst)
	}()

	fmt.Printf("%+v\n", data)
	first, err := moviego.Load(src)

	if err != nil {
		log.Println(err)
		return
	}

	err = first.SubClip(data.Start, data.End).Output(dst).Run()
	if err != nil {
		log.Fatal(err)
		return
	}

	go uploadFile(bucketName, dst, dst)

	outFile := fmt.Sprintf("gs://%s/%s", bucketName, dst)
	log.Println("out", outFile)

	err = register2DB(data)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("entry has been registered to DB")

	msg.Ack()
}

func register2DB(data params) error {

	ctx := context.Background()
	client, _ := firestore.NewClient(ctx, projectID)
	doc := data
	client.Collection("data").Doc(data.UserID).Set(ctx, doc)
	return nil
}

func downloadFile(bucket, object string, destFileName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	f, err := os.Create(destFileName)
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

	// ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	// defer cancel()

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
