package eml

import (
	"context"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

func (e *emlStore) onBeforeRequest(_ *resty.Client, req *resty.Request) error {
	// Skip for /token request as has own auth
	if strings.Contains(req.URL, pathToken) {
		return nil
	}
	if e.token.ShouldRefresh() {
		if e.token.IsValid() {
			e.refreshTokenAsync(req.Context())
		} else if err := e.refreshToken(req.Context()); err != nil {
			return ContextualError(err, "eml.refreshToken")
		}
	}
	req.SetAuthScheme("Bearer")
	req.SetAuthToken(e.token.Value)
	return nil
}

func logRequest(_ *resty.Client, req *resty.Request) error {
	log.Println("EML", req.Method, "Length", req.URL, "->", req.URL)
	return nil
}

func logResponse(_ *resty.Client, res *resty.Response) error {
	log.Println("EML", res.Status(), "Length", res.Size(), "in", res.Time(), "<-", res.Request.URL)
	if !res.IsSuccess() {
		log.Println(string(res.Body()))
	}
	return nil
}

func checkError(resp *resty.Response, err error) error {
	if err != nil {
		return ContextualError(err, "resty")
	}
	if !resp.IsSuccess() {
		statusCode := resp.StatusCode()
		err := resp.Error().(*ErrorModel)
		if statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound {
			return NewHttpError(statusCode, err, err)
		}
		return ContextualError(err, "eml status %d", statusCode)
	}
	return nil
}

func (e *emlStore) refreshTokenAsync(ctx context.Context) {
	if atomic.CompareAndSwapUint32(&e.refreshing, 0, 1) {
		go func() {
			err := e.refreshToken(ctx)
			atomic.StoreUint32(&e.refreshing, 0)
			if err != nil {
				log.Println("async token refresh error:", err)
			}
		}()
	}
}

func (e *emlStore) refreshToken(ctx context.Context) error {
	log.Println("Refreshing EML access token")
	resp, err := e.request(ctx).
		SetFormData(map[string]string{"grant_type": "client_credentials"}).
		SetResult(&TokenResponse{}).
		Post(pathToken)
	if err != nil {
		log.Printf("Error Refreshing EML access token : %v", err)
		return err
	}
	if !resp.IsSuccess() {
		log.Printf("Token Retrieval Unsuccessful : %v", resp.Error())
		return resp.Error().(*ErrorModel)
	}
	log.Println("Token Updates")
	e.token = mapBearerToken(resp.Result().(*TokenResponse))
	return nil
}
