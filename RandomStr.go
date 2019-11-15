package main

import (
        "fmt"
        "math/rand"
        "time"
)

func main() {
        fmt.Println(RandString(16))
}

var source = rand.NewSource(time.Now().UnixNano())

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-=_+[]{};:,.?<>"

func RandString(length int) string {
        b := make([]byte, length)
        for i := range b {
                b[i] = charset[source.Int63()%int64(len(charset))]
        }
        return string(b)
}
