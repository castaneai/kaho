package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/image"

	"io"
	"net/http"

	"google.golang.org/appengine"
)

const (
	bucketName = "libeccio-kaho"
)

type UploadResult struct {
	BucketName string
	ObjectName string
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	appengine.Main()
}

func handleError(w http.ResponseWriter, err error) {
	log.Printf("[err] %+v", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("error"))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	f, _, err := r.FormFile("file")
	if err != nil {
		handleError(w, err)
		return
	}
	defer f.Close()

	if appengine.IsDevAppServer() {
		handleError(w, errors.New("kaho cannot run on development server"))
		return
	}

	ctx := r.Context()
	res, err := upload(ctx, f)
	if err != nil {
		handleError(w, err)
		return
	}

	servingURL, err := generateServingUrl(ctx, res, true)
	if err != nil {
		handleError(w, err)
		return
	}

	log.Printf("success upload: %s", servingURL)
	w.Header().Set("Location", servingURL)
	w.WriteHeader(http.StatusSeeOther)
}

func generateServingUrl(ctx context.Context, result *UploadResult, isSecureURL bool) (string, error) {
	// NOTE: https://cloud.google.com/appengine/docs/standard/go/blobstore/reference#BlobKeyForFile
	gsURL := fmt.Sprintf("/gs/%s/%s", result.BucketName, result.ObjectName)
	blobKey, err := blobstore.BlobKeyForFile(ctx, gsURL)
	if err != nil {
		return "", err
	}
	servingURLOpts := &image.ServingURLOptions{Secure: isSecureURL}
	url, err := image.ServingURL(ctx, blobKey, servingURLOpts)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func upload(ctx context.Context, f io.ReadSeeker) (*UploadResult, error) {
	fileID, err := generateFileID()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 512)
	n, err := io.ReadAtLeast(f, buf, 1)
	if err != nil {
		return nil, err
	}
	contentType := http.DetectContentType(buf[:n])
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	if err := uploadToCloudStorage(ctx, bucketName, fileID, f, contentType); err != nil {
		return nil, err
	}
	return &UploadResult{
		BucketName: bucketName,
		ObjectName: fileID,
	}, nil
}

func generateFileID() (string, error) {
	uu, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uu.String(), nil
}

func uploadToCloudStorage(ctx context.Context, bucketName, objName string, r io.Reader, contentType string) error {
	st, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	b := st.Bucket(bucketName)
	obj := b.Object(objName)
	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
