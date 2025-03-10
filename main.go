package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("running at http://localhost:3000")
	panic(http.ListenAndServe(":3000", nil))
}
