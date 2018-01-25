package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/image"
	"google.golang.org/appengine/log"
	"io"
	"net/http"
)

const (
	bucketName = "cafe-stile"
)

type UploadResult struct {
	BucketName string
	ObjectName string
}

func init() {
	r := gin.Default()
	r.POST("/upload", uploadHandler)
	http.Handle("/", r)
}

func handleError(ctx context.Context, gin *gin.Context, err error) {
	log.Errorf(ctx, "[err] %+v", errors.WithStack(err))
	gin.AbortWithError(http.StatusInternalServerError, err)
}

func uploadHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	f, err := c.FormFile("file")
	if err != nil {
	}

	r, err := f.Open()
	if err != nil {
		handleError(ctx, c, err)
		return
	}

	if appengine.IsDevAppServer() {
		handleError(ctx, c, errors.New("kaho cannot run on development server"))
		return
	}

	res, err := upload(ctx, r)
	if err != nil {
		handleError(ctx, c, err)
		return
	}

	servingURL, err := generateServingUrl(ctx, res, true)
	if err != nil {
		handleError(ctx, c, err)
		return
	}

	log.Infof(ctx, "success upload: %s", servingURL)
	c.Redirect(http.StatusSeeOther, servingURL)
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
	uu, err := uuid.NewV4()
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
