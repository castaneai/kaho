package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

var redirectErr = errors.New("redirect")

func UploadToKaho(ctx context.Context, kahoURL string, filename string, body io.Reader) (*url.URL, error) {
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	w, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(w, body); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", kahoURL+"/upload", &b)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.ContentLength = int64(b.Len())
	req.Header.Set("Content-Type", mw.FormDataContentType())

	hc := &http.Client{CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return redirectErr
	}}
	res, err := hc.Do(req)
	if res == nil {
		return nil, err
	}
	defer res.Body.Close()
	if err != nil && err != redirectErr {
		rb, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("kaho should respond status: 302 See Other, but %v found (body: %v)", res.Status, string(rb))
	}
	newURL, err := res.Location()
	if err != nil {
		return nil, err
	}
	return newURL, nil
}
