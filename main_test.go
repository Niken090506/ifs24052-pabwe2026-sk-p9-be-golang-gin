package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainFunc(t *testing.T) {
	called := false
	err := runServer(":8081", func(addr ...string) error {
		called = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, called)

	orig := runServer
	defer func() { runServer = orig }()
	var gotAddr string
	runServer = func(addr string, run func(...string) error) error {
		gotAddr = addr
		return nil
	}
	os.Setenv("DB_CONNECTION", "sqlite")
	defer os.Unsetenv("DB_CONNECTION")
	main()
	assert.NotEmpty(t, gotAddr)
}
