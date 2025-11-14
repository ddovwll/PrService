package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	qwe := struct {
		A string `json:"a"`
	}{}

	err := json.Unmarshal([]byte(`{"a":1}`), &qwe)
	fmt.Println(err)
	fmt.Println(qwe)
}
