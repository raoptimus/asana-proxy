package asana

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type (
	Options struct {
		URL        string
		ServerAddr string
	}
	Proxy struct {
		sync.Mutex
		options Options
		signal  chan os.Signal
		cache   map[string]ResponseData
	}
	ResponseData struct {
		url        url.URL
		headers    http.Header
		statusCode int
		data       []byte
	}
)

func NewProxy(options Options) *Proxy {
	return &Proxy{
		options: options,
		signal:  make(chan os.Signal),
		cache:   make(map[string]ResponseData),
	}
}

func (s *Proxy) Run() error {
	go s.background()
	go s.backgroundCacheClear()

	signal.Notify(s.signal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-s.signal:
		log.Infof("proxy has received shutdown signal [%s]", sig)
	}

	return nil
}
