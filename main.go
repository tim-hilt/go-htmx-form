package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed index.html templates/*
var f embed.FS

var m sync.Mutex

type Counter struct {
	CounterValue int64
}

func (c *Counter) Increment() {
	m.Lock()
	defer m.Unlock()
	c.CounterValue++
}

func (c *Counter) Decrement() {
	m.Lock()
	defer m.Unlock()
	c.CounterValue--
}

func (c *Counter) Reset() {
	m.Lock()
	defer m.Unlock()
	c.CounterValue = 0
}

func main() {
	counter := Counter{}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	addCounterEndpoint := func(endpoint string, modifyCounter func()) {
		r.POST(endpoint, func(c *gin.Context) {
			modifyCounter()
			c.HTML(http.StatusOK, "templates/counter.tmpl", counter)
		})
	}

	r.SetTrustedProxies([]string{})

	ts, err := template.ParseFS(f, "index.html", "templates/counter.tmpl")

	if err != nil {
		os.Exit(1)
	}

	r.SetHTMLTemplate(ts)

	// INFO: This is good for quick prototyping, because refreshing the
	//  browser refreshes the index.html
	// r.StaticFile("/", "./index.html")

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", counter)
	})

	addCounterEndpoint("/increment", counter.Increment)
	addCounterEndpoint("/decrement", counter.Decrement)
	addCounterEndpoint("/reset", counter.Reset)

	r.Run()
}
