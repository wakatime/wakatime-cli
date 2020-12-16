package main

// list every possible way of importing to try and break dependency detection
// http://learngowith.me/alternate-ways-of-importing-packages/

import "fmt"
import "compress/gzip"
import "github.com/golang/example/stringutil"
import (
	"log"
	"os"
)
import newname "oldname"
import . "direct"
import _ "suppress"
import (
	"foobar"
	_ "image/gif"
	. "math"
)

func main() {
	fmt.Println("Hello, World!")
}
