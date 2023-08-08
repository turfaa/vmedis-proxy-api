package client

import (
	"context"
	"testing"

	"golang.org/x/time/rate"
)

func TestX(t *testing.T) {
	c := New(
		"https://aulia.vmedis.com",
		[]string{"ob9g1kfdk8aev1cb2nr76qkcj6"},
		1,
		rate.NewLimiter(100, 100),
	)

	d, err := c.GetDrug(context.Background(), 2094)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", d)
}
