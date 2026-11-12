package vmedisv2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Client calls the vmedis report API, the one behind the "-v2" report pages
// (served from report.vmedis.com), through the gateway, for one apotek.
type Client struct {
	gateway gatewayCaller
	appID   string
	reg     string
}

type gatewayCaller interface {
	Call(ctx context.Context, req GatewayRequest, responseTarget any) error
}

// New creates a Client for the apotek identified by appID (the "a" and
// "login_a" parameters the report pages send, e.g. "v7965") and reg (the
// "reg" parameter, e.g. "db8c").
func New(gateway *GatewayClient, appID, reg string) *Client {
	return &Client{
		gateway: gateway,
		appID:   appID,
		reg:     reg,
	}
}

// reportApiVersion is the Target-Version the report pages call the report
// API with.
const reportApiVersion = "apivi"

// callReport POSTs payload, plus the parameters identifying the apotek, to
// the report API endpoint at targetUrl.
func (c *Client) callReport(ctx context.Context, targetUrl string, payload map[string]any, responseTarget any) error {
	if c.appID == "" || c.reg == "" {
		return fmt.Errorf("vmedis v2 app id and reg must be configured")
	}

	payloadJson, _ := json.Marshal(payload)
	log.Printf("POST %s %s", targetUrl, payloadJson)

	params := make(map[string]any, len(payload)+3)
	for k, v := range payload {
		params[k] = v
	}
	params["a"] = c.appID
	params["login_a"] = c.appID
	params["reg"] = c.reg

	return c.gateway.Call(ctx, GatewayRequest{
		TargetUrl:     targetUrl,
		TargetVersion: reportApiVersion,
		TargetOptions: GatewayTargetOptions{Method: "POST"},
		Payload:       params,
	}, responseTarget)
}
