package client_test

import (
	"os"
	"stickyants-go-sharepoint/client"
	"testing"

	"github.com/joho/godotenv"
)

func TestSPClientConnect(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal(err, "Please provide .env file with SPURL, SPAppId, SPAppSecret variables")
	}

	c := client.NewSPClient(os.Getenv("SPURL"),
		os.Getenv("SPAppId"),
		os.Getenv("SPAppSecret"))
	c.Dev()

	data := c.Get("Lists/GetByTitle('Documents')")
	t.Log(data)

}
