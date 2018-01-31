package gateway

import (
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	rateLimitHandler "github.com/devopsfaith/krakend-ratelimit/juju/router/gin"
	router "github.com/devopsfaith/krakend/router/gin"

	"context"
	"fmt"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/gin-gonic/gin"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	"net/http"
	"strings"
	"time"
)

func NewHandlerFactory(metricCollector *metrics.Metrics) router.HandlerFactory {
	handlerFactory := rateLimitHandler.NewRateLimiterMw(BaseHandlerFactory)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	return handlerFactory
}

func BaseHandlerFactory(endpointConfig *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
	endpointTimeout := time.Duration(endpointConfig.Timeout) * time.Millisecond
	cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(endpointConfig.CacheTTL.Seconds()))
	isCacheEnabled := endpointConfig.CacheTTL.Seconds() != 0
	emptyResponse := gin.H{}

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, endpointTimeout)

		//c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)
		response, err := proxy(requestCtx, NewRequest(c, endpointConfig.QueryString))
		for _, k := range HeadersToReturn {
			if h, ok := response.Metadata.Headers[k]; ok {
				c.Header(k, h[0])
			}
		}

		if err != nil {
			c.AbortWithStatusJSON(ToHTTPError(err), gin.H{
				core.ErrorResponseReturnKey: core.CreateError(core.GeneralError, err.Error()),
			})
			cancel()
			return
		}

		select {
		case <-requestCtx.Done():
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				core.ErrorResponseReturnKey: core.CreateError(core.GeneralError, requestCtx.Err().Error()),
			})
			cancel()
			return
		default:
		}

		if isCacheEnabled && response != nil && response.IsComplete {
			c.Header("Cache-Control", cacheControlHeaderValue)
		}

		if response == nil {
			c.JSON(http.StatusOK, gin.H{
				core.ErrorResponseReturnKey: emptyResponse,
			})
			cancel()
			return
		}

		_, hasError := response.Data[core.ErrorResponseMessageKey]
		var responseKey string
		var status int
		if response.Metadata.StatusCode >= 400 {
			responseKey = core.ErrorResponseReturnKey
			status = response.Metadata.StatusCode
		} else if hasError {
			responseKey = core.ErrorResponseReturnKey
			status = http.StatusOK
		} else {
			responseKey = core.ResponseReturnKey
			status = response.Metadata.StatusCode
		}
		c.JSON(status, gin.H{
			responseKey: response.Data,
		})
		cancel()
	}
}

var (
	HeadersToSend   = []string{"Content-Type", "Accept-Language", core.ClientID, core.ClientEmail, core.TracingRoot, core.TracingPath}
	HeadersToReturn = []string{"Content-Type", core.TracingRoot, core.TracingPath}
)

func NewRequest(c *gin.Context, queryString []string) *proxy.Request {
	params := make(map[string]string, len(c.Params))
	for _, param := range c.Params {
		params[strings.Title(param.Key)] = param.Value
	}

	headers := make(map[string][]string, 2+len(HeadersToSend))
	headers["X-Forwarded-For"] = []string{c.ClientIP()}

	for _, k := range HeadersToSend {
		if h, ok := c.Request.Header[k]; ok {
			headers[k] = h
		}
	}

	query := make(map[string][]string, len(queryString))
	for i := range queryString {
		if v := c.Request.URL.Query().Get(queryString[i]); v != "" {
			query[queryString[i]] = []string{v}
		}
	}

	return &proxy.Request{
		Method:  c.Request.Method,
		Query:   query,
		Body:    c.Request.Body,
		Params:  params,
		Headers: headers,
	}
}

func ToHTTPError(_ error) int {
	return http.StatusInternalServerError
}
