package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

const (
	listenOn = 8101
	proxyAt  = 8100

	keyHeader = "X-WebXG-Proc-Key"
)

var c = fasthttp.Client{}

func main() {

	fmt.Println("Listening on ", listenOn)
	fmt.Println("Upstreaming to", proxyAt)

	err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", listenOn), requestHandler)

	if err != nil {
		panic(err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	pqk := ctx.Request.Header.Peek(keyHeader)

	if len(pqk) > 0 {
		// Existing check on rendering a request already seen

		// long tailed asymptotic distribution - Mostly responses 100ms to 1000ms.
		toAck := 1000 / ((float64(rand.Intn(9)+1) * float64(0.8)) + float64(0.2))

		time.Sleep(time.Duration(toAck) * time.Millisecond)

		fmt.Fprintf(ctx, "You again? Okay, I've responded to your uuid ( %s ) in %.2f seconds\n", pqk, toAck)
		fmt.Printf("3rd > %.2f\n", toAck)

	} else {
		// A fresh request. Give them a new id back, and later tell them it's ready

		// long tailed asymptotic distribution - Mostly responses 100ms to 1000ms.
		toAck := 1000 / ((float64(rand.Intn(9)+1) * float64(0.8)) + float64(0.2))

		// long tailed distribution between 600ms and 8000ms, mostly
		toReady := (8000 / ((float64(rand.Intn(9)+1) * float64(0.8)) + float64(0.2))) - 400

		time.Sleep(time.Duration(toAck) * time.Millisecond)

		fmt.Fprintf(ctx, "Hi there, newcomer! I've delayed you by %.2f, but Your response will be ready in %.2f\n", toAck, toReady)
		fmt.Printf("2nd > %.2f\n", toAck)

		id := uuid.NewV4()
		ctx.Response.Header.Set(keyHeader, id.String())
		ctx.Response.SetStatusCode(102)

		go readyLater(id, toReady)
	}
}

func readyLater(id uuid.UUID, delay float64) {

	time.Sleep(time.Duration(delay) * time.Millisecond)

	req := &fasthttp.Request{}

	req.Header.SetMethod("READY")
	req.Header.SetRequestURI(fmt.Sprintf("http://localhost:%d/%s", proxyAt, id.String()))

	fmt.Printf("2nd > %.2f\n", delay)
	resp := &fasthttp.Response{}

	err := c.Do(req, resp)
	if err != nil {
		panic(err)
	}
}
