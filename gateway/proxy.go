package gateway

import (
	circuitBreaker "github.com/devopsfaith/krakend-circuitbreaker/gobreaker/proxy"
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	rateLimitProxy "github.com/devopsfaith/krakend-ratelimit/juju/proxy"
	"github.com/devopsfaith/krakend/proxy"

	"context"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"net/http"
)

func NewProxy(logger logging.Logger, metricCollector *metrics.Metrics) proxy.Factory {
	// create the metrics collector
	metricEnabledProxy := proxy.NewDefaultFactory(BackendFactory(metricCollector), logger)
	metricEnabledProxy = metricCollector.ProxyFactory("pipe", metricEnabledProxy)
	return metricEnabledProxy
}

func BackendFactory(metricCollector *metrics.Metrics) proxy.BackendFactory {
	backendFactory := rateLimitProxy.BackendFactory(CustomHTTPProxyFactory())
	backendFactory = circuitBreaker.BackendFactory(backendFactory)
	backendFactory = metricCollector.BackendFactory("backend", backendFactory)
	return backendFactory
}

// CustomHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func CustomHTTPProxyFactory() proxy.BackendFactory {
	return func(backend *config.Backend) proxy.Proxy {
		ef := proxy.NewEntityFormatter(backend.Target, backend.Whitelist, backend.Blacklist, backend.Group, backend.Mapping)
		rp := HTTPResponseParserFactory(proxy.HTTPResponseParserConfig{Decoder: backend.Decoder, EntityFormatter: ef})
		return proxy.NewHTTPProxyDetailed(backend, proxy.DefaultHTTPRequestExecutor(proxy.NewHTTPClient), PassThroughHTTPStatusHandler, rp)
	}
}

// DefaultHTTPCodeHandler is the default implementation of HTTPStatusHandler
func PassThroughHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

// DefaultHTTPResponseParserFactory is the default implementation of HTTPResponseParserFactory
func HTTPResponseParserFactory(cfg proxy.HTTPResponseParserConfig) proxy.HTTPResponseParser {
	return func(ctx context.Context, resp *http.Response) (*proxy.Response, error) {
		var data map[string]interface{}
		err := cfg.Decoder(resp.Body, &data)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		newResponse := proxy.Response{Metadata: proxy.Metadata{StatusCode: resp.StatusCode, Headers: resp.Header}, Data: data, IsComplete: true}
		newResponse = cfg.EntityFormatter.Format(newResponse)
		return &newResponse, nil
	}
}
