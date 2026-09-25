package main

import (
	"context"
	"fmt"
	client "github.com/explainableaixai/shadowaitools-go"
	"os"
)

func main() {
	c := client.New(os.Getenv("AQ_API_KEY"))
	result, err := c.Scan(context.Background(), "dns-export.csv")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result)
}
