package main

import (
	"fmt"
	"net/url"
	"path"
	"testing"
)

func Test_main(t *testing.T) {
	host := "http://106.15.79.230:6789/core-metadata/"
	newUrl, err := url.Parse(host)
	fmt.Println(newUrl, err)

	fmt.Println(newUrl.Host, newUrl.Hostname(), newUrl.Path, newUrl.Port(), newUrl.Scheme)

	fmt.Println(path.Clean("/core-meta"))
	fmt.Println(path.Clean("//xx"))
	fmt.Println(path.Clean("/path/:id/test"))
	fmt.Println(path.Clean("/////"))
	fmt.Println(path.Clean(""))
	fmt.Println(path.Clean("*"))
}
