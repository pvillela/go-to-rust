package main

import "os"

var src0 = `
package p
const c = 1.0
var X = f(3.14)*2 + c
`

var src1 = (func() string {
	src, err := os.ReadFile("out/article.go")
	if err != nil {
		panic(err)
	}
	return string(src)
})()
