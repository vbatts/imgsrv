package main

import (
	"fmt"
)

var (
	MAJOR   int    = 1
	MINOR   int    = 0
	TINY    int    = 0
	VERSION string = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, TINY)
)
