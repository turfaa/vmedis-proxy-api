package vmedisv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type GatewayClient struct {
	url        string
	httpClient *http.Client
	crypt      Crypt
}

type GatewayRequest struct {
	TargetUrl     string
	TargetVersion string
	TargetOptions GatewayTargetOptions
	Payload       any
}

type GatewayTargetOptions struct {
	// Method defaults to GET if not set.
	Method string
}

func (o GatewayTargetOptions) String() string {
	method := strings.ToUpper(o.Method)
	if method == "" {
		method = "GET"
	}

	return fmt.Sprintf(`{"method":"%s"}`, method)
}

func NewGatewayClient(baseUrl string, crypt *Crypt) *GatewayClient {
	return &GatewayClient{
		url:        baseUrl,
		httpClient: &http.Client{Timeout: time.Minute},
		crypt:      *crypt,
	}
}

func (c *GatewayClient) Call(ctx context.Context, req GatewayRequest, responseTarget any) error {
	httpReq, err := c.buildHTTPRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("build HTTP request: %w", err)
	}

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do HTTP request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if err := c.parseHTTPResponseBody(bodyBytes, responseTarget); err != nil {
		return fmt.Errorf("parse HTTP response body: %w", err)
	}

	return nil
}

func (c *GatewayClient) buildHTTPRequest(ctx context.Context, req GatewayRequest) (*http.Request, error) {
	body, err := c.buildHTTPRequestBody(req)
	if err != nil {
		return nil, fmt.Errorf("build HTTP body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if err := c.appendHeadersToHTTPReq(httpReq, req); err != nil {
		return nil, fmt.Errorf("append headers to HTTP request: %w", err)
	}

	return httpReq, nil
}

func (c *GatewayClient) buildHTTPRequestBody(req GatewayRequest) ([]byte, error) {
	payload, err := c.crypt.Encrypt(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}

	body := struct {
		Params string `json:"params"`
	}{
		Params: string(payload),
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	return bodyBytes, nil
}

func (c *GatewayClient) appendHeadersToHTTPReq(httpReq *http.Request, gatewayReq GatewayRequest) error {
	httpReq.Header.Add("Content-Type", "application/json")

	targetUrl, err := c.crypt.EncryptToURLEncoded(gatewayReq.TargetUrl)
	if err != nil {
		return fmt.Errorf("encrypt target URL: %w", err)
	}
	httpReq.Header.Add("Target-Url", targetUrl)

	targetVersion, err := c.crypt.EncryptToURLEncoded(gatewayReq.TargetVersion)
	if err != nil {
		return fmt.Errorf("encrypt target version: %w", err)
	}
	httpReq.Header.Add("Target-Version", targetVersion)

	targetOptions, err := c.crypt.EncryptToURLEncoded(gatewayReq.TargetOptions.String())
	if err != nil {
		return fmt.Errorf("encrypt target options: %w", err)
	}
	httpReq.Header.Add("Target-Options", targetOptions)

	return nil
}

func (c *GatewayClient) parseHTTPResponseBody(bodyBytes []byte, responseTarget any) error {
	body := struct {
		Error string `json:"error"`
		Data  string `json:"data"`
	}{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		log.Printf("failed to unmarshal response body: %s", string(bodyBytes))
		return fmt.Errorf("unmarshal response body: %w", err)
	}

	if body.Error != "" {
		return fmt.Errorf("gateway error: %s", body.Error)
	}

	if err := c.crypt.Decrypt([]byte(body.Data), responseTarget); err != nil {
		return fmt.Errorf("decrypt response body to %T: %w", responseTarget, err)
	}

	return nil
}
