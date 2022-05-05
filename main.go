package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

// go run . && cat output.txt | grep :true

type Target struct {
	InputFilePath  string // Путь до файла с паролями
	OutputFilePath string // Путь до файла, куда должны записываться результаты
	*Connection
}

//solution 1
func HackServer(ctx context.Context, target *Target) {
	b, err := ioutil.ReadFile(target.InputFilePath) // just pass the file name
	if err != nil {
		fmt.Print(err)
		return
	}

	res := make(chan string, 10000)
	var wg sync.WaitGroup
	var str []string = strings.Split(string(b), "\n")
	output, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer output.Close()
	for _, password := range str {
		wg.Add(1)
		res <- password
		go func() {
			defer wg.Done()
			pass, ok := <-res
			if !ok {
				return
			}
			req := &Request{
				Ctx:      ctx,
				Password: pass,
			}
			SendRequest(target.Connection, req)
			resp, ok := <-target.Connection.ResponseConn
			if !ok {
				return
			}
			if resp.Pass {
				output.Write([]byte(resp.Password + ":true\n"))
			} else {
				output.Write([]byte(resp.Password + ":false\n"))
			}
		}()

	}
	wg.Wait()
}

//solution 2

// func GetPasswordsFromFile(path string) []string {
// 	data, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		log.Println(err)
// 		return []string{}
// 	}
// 	return strings.Split(string(data), "\n")
// }

// func HackServer(ctx context.Context, target *Target) {
// 	log.SetFlags(log.Llongfile)

// 	file, err := os.OpenFile(target.OutputFilePath, os.O_CREATE|os.O_RDWR, 0755)
// 	if err != nil {
// 		log.Println(err)
// 		os.Exit(1)
// 	}
// 	defer file.Close()

// 	passwords := GetPasswordsFromFile(target.InputFilePath)
// 	var wg sync.WaitGroup
// 	for i := range passwords {
// 		wg.Add(1)

// 		go func(idx int) {
// 			defer wg.Done()
// 			SendRequest(target.Connection, &Request{ctx, passwords[idx]})
// 			resp := <-target.Connection.ResponseConn

// 			suffix := "false"
// 			if resp.Pass {
// 				suffix = "true"
// 			}
// 			fmt.Fprintf(file, "%s:%s\n", resp.Password, suffix)
// 		}(i)
// 	}
// 	wg.Wait()
// }

func main() {
	requestChan := make(chan *Request)
	responseChan := make(chan *Response)
	defer close(requestChan)
	defer close(responseChan)

	connection := &Connection{
		RequestConn:  requestChan,
		ResponseConn: responseChan,
	}

	target := &Target{
		InputFilePath:  "darkweb2017-top10000.txt",
		OutputFilePath: "output.txt",
		Connection:     connection,
	}
	// Заменить "Password" на один из 10000 паролей
	server := NewVulnerableServer("Password", connection)

	go server.Run()

	// // Пробовать запускать с разными контекстами
	ctx := context.Background()
	HackServer(ctx, target)
}
