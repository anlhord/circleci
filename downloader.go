package main

import "net/http"
import "time"
import "fmt"
import "encoding/json"
import "io"
import "crypto/sha256"
//import "net/url"
//import "strings"
import "bytes"


type Submission struct {
	Ts  string			`json:"ts"`
	Event string			`json:"event"`
	Collection string		`json:"collection"`
	Data struct {
		Version  string		`json:"version"`
		Body  string		`json:"body"`
		Compiler  string	`json:"compiler"`
	}
}


func sha(body string) []byte {
	h := sha256.New()
	io.WriteString(h, body)
	return h.Sum(nil)
}

func nodata(buf []byte) []byte {
	if len(buf) <= 6 {
		return buf
	}
	if buf[0] == 'd' && buf[1] == 'a' && buf[2] == 't' && buf[3] == 'a' &&
		buf[4] == ':' && buf[5] == ' ' {
		return buf[6:]
	}
	return buf
}

func line2(buf []byte) []byte {
	for i := 2; i > 0; buf = buf[1:] {
		if buf[0] == '\n' {
			i--
		}
	}
	for i := 0; i <  len(buf); i++ {
		if buf[i] == '\n' {
			buf = buf[:i]
		}
	}
	return buf
}

func main() {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:  24* 60 * 60 * time.Second,
//		DisableCompression: true,
	}


	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", "https://queue1-83fb.restdb.io/realtime", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-apikey", "59cfeb7504067cfd77ad9b8d")


	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}


	_ = resp

	var submission []byte

	for {

	var buf [32]byte
	var extra [32]byte

	l, err := resp.Body.Read(buf[:])

	if err != nil {
		panic(err.Error())
	}

	if buf[0] == 'i' && buf[1] == 'd' && buf[2] == ':' {

	submission = buf[:l]

	for l >= 32 {
		l, err = resp.Body.Read(extra[:])
		submission = append(submission, extra[:l]...)
	}

	submission = nodata(line2(submission))

	var object Submission

	err := json.Unmarshal(submission, &object)
	if err != nil {
		fmt.Println(err.Error())
		continue
	}

	var hash = sha(object.Data.Body)

	var hash64 = fmt.Sprintf("%x", hash)
	fmt.Printf("Received %s\n", hash64)

	response, err := compileAndRun(nil, &object)
	if err != nil {
		fmt.Println(err.Error())
		continue
	}

	response.Hash = hash64

	jsoned, err3 := json.Marshal(response)
	if err3 != nil {
		fmt.Println(err3.Error())
		continue
	}
	jsoned = append([]byte(`{"Response":`), jsoned...)
	jsoned = append(jsoned, '}')


//////// send http put

	{

    url := "https://queue2-9029.restdb.io/rest/response/59d7dbc8a10f1169000795"+hash64[:2]
    fmt.Println("URL:>", url)

    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsoned))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-apikey", "59d7b8f016d89bb7783291a4")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }

    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
//    body, _ := ioutil.ReadAll(resp.Body)
//    fmt.Println("response Body:", string(body))


    resp.Body.Close()

	}




//	fmt.Println(object.Data.Body)

	}

	}


}
