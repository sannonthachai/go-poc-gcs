package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load()
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/save", save)

	e.Logger.Fatal(e.Start("localhost:1323"))
}

func save(c echo.Context) error {
	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get avatar
	avatar, err := c.FormFile("image")
	if err != nil {
		return err
	}

	// Source
	src, err := avatar.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	object := "test/image/" + avatar.Filename

	uploadImageGCS(client, "test2-image-sannon", object, src)

	return c.HTML(http.StatusOK, "OK")
}

func createBucket(client *storage.Client, name string) error {
	// Sets your Google Cloud Platform project ID.
	projectID := os.Getenv("PROJECT_ID")
	ctx := context.Background()

	// Creates a Bucket instance.
	bucket := client.Bucket(name)

	// Creates the new bucket.
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := bucket.Create(ctx, projectID, &storage.BucketAttrs{
		StorageClass: "STANDARD",
		Location:     "asia",
	}); err != nil {
		return err
	}

	return nil
}

func uploadImageGCS(client *storage.Client, bucket, object string, fp multipart.File) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := client.Bucket(bucket).Attrs(ctx)
	if err != nil {
		fmt.Println(err)

		if err := createBucket(client, bucket); err != nil {
			fmt.Println(err)
		}
	}

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(wc, fp); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
