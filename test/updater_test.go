package test

import (
	"os"
	"testing"
)

var fetcher = Fetcher{}

func TestFetcher(*testing.T) {
	_, err := fetcher.Pull()
	if err != nil {
		panic(err)
	}
}

func BenchmarkFetcher(*testing.B) {
	_, err := fetcher.Pull()
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
