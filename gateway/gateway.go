package gateway

import (
	"context"
	"github.com/devopsfaith/krakend-gologging"
	metrics "github.com/devopsfaith/krakend-metrics/gin"
	"github.com/devopsfaith/krakend-viper"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	router "github.com/devopsfaith/krakend/router/gin"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case sig := <-sigs:
			log.Println("Signal intercepted:", sig)
			cancel()
		case <-ctx.Done():
			log.Fatal(ctx.Err())
		}
	}()
	krakendConfig, err := viper.New().Parse("krakend.json")
	if err != nil {
		panic(err)
	}
	NewExecutor(krakendConfig, ctx)
}

func NewExecutor(cfg config.ServiceConfig, ctx context.Context) {

	logger, err := gologging.NewLogger(cfg.ExtraConfig)
	if err != nil {
		logger, err = logging.NewLogger("DEBUG", os.Stdout, "[Cryptopia]")
		if err != nil {
			return
		}
		logger.Error("unable to create the gologgin logger:", err.Error())
	}

	// create the metrics collector
	metricCollector := metrics.New(ctx, time.Minute, logger)

	// setup the krakend router
	routerFactory := router.NewFactory(router.Config{
		Engine:         NewEngine(metricCollector),
		ProxyFactory:   NewProxy(logger, metricCollector),
		Middlewares:    []gin.HandlerFunc{NewAuthenticationEnabledMiddleware(cfg), AddTracingMiddleware()},
		Logger:         logger,
		HandlerFactory: NewHandlerFactory(metricCollector),
	})

	// start the engines
	routerFactory.NewWithContext(ctx).Run(cfg)

}

func NewEngine(metricCollector *metrics.Metrics) *gin.Engine {
	engine := gin.Default()
	engine.GET("/__stats/", metricCollector.NewExpHandler())

	return engine
}
