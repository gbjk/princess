package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/satori/go.uuid"
	"github.com/streamrail/concurrent-map"
	"github.com/valyala/fasthttp"
)

const (
	listenOn = 8101
	proxyAt  = 8100

	keyHeader = "X-WebXG-Proc-Key"
	idHeader  = "X-WebXG-Request-ID"
)

var c = fasthttp.Client{}
var r *rand.Rand
var requests cmap.ConcurrentMap

type requestTimer struct {
	id     string
	phase1 float64
	phase2 float64
	phase3 float64
	total  float64
}

func main() {

	fmt.Println("Listening on ", listenOn)
	fmt.Println("Upstreaming to", proxyAt)

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	requests = cmap.New()

	err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", listenOn), requestHandler)

	if err != nil {
		panic(err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	pqk := ctx.Request.Header.Peek(keyHeader)
	rId := string(ctx.Request.Header.Peek(idHeader))

	if len(pqk) > 0 {
		// Existing check on rendering a request already seen

		// long tailed asymptotic distribution - Mostly responses 100ms to 1000ms.
		toAck := 1000 / ((((r.Float64() * 9) + 1) * float64(0.8)) + float64(0.2))

		time.Sleep(time.Duration(toAck) * time.Millisecond)

		timerI, exists := requests.Get(rId)
		if !exists {
			panic("Really expected everything with a pqk to be in our map")
		}
		timer := timerI.(requestTimer)

		timer.phase3 = toAck
		timer.total = timer.phase1 + timer.phase2 + timer.phase3

		requests.Remove(rId)

		ctx.Response.Header.Set(idHeader, rId)
		fmt.Fprintf(ctx, "%.9f", timer.total)

	} else {
		// A fresh request. Give them a new id back, and later tell them it's ready

		// long tailed asymptotic distribution - Mostly responses 100ms to 1000ms.
		toAck := 1000 / ((((r.Float64() * 9) + 1) * float64(0.8)) + float64(0.2))

		// long tailed distribution between 600ms and 8000ms, mostly
		toReady := (8000 / ((((r.Float64() * 9) + 1) * float64(0.8)) + float64(0.2))) - 400

		timer := requestTimer{
			id:     rId,
			phase1: toAck,
			phase2: toReady,
		}
		requests.Set(rId, timer)

		time.Sleep(time.Duration(toAck) * time.Millisecond)

		fmt.Fprintf(ctx, "Hi there, newcomer! I've delayed you by %.2f, but Your response will be ready in %.2f\n", toAck, toReady)

		id := uuid.NewV4()
		ctx.Response.Header.Set(keyHeader, id.String())
		ctx.Response.Header.Set(idHeader, rId)
		ctx.Response.SetStatusCode(102)

		go readyLater(id, rId, toReady)
	}
}

func readyLater(id uuid.UUID, rId string, delay float64) {

	time.Sleep(time.Duration(delay) * time.Millisecond)

	req := &fasthttp.Request{}

	req.Header.SetMethod("READY")
	req.Header.SetRequestURI(fmt.Sprintf("http://localhost:%d/%s", proxyAt, id.String()))
	req.Header.Set(idHeader, rId)

	resp := &fasthttp.Response{}

	err := c.Do(req, resp)
	if err != nil {
		panic(err)
	}
}
