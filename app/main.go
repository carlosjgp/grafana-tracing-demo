package main

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"

	"google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	srvAddr     = ":8080"
	metricsAddr = ":9090"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	fn := initTracer()
	defer fn()

	tr := global.Tracer("server")

	// Create our middleware.
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// Create our server.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		attrs, entries, spanCtx := httptrace.Extract(req.Context(), req)

		rnd := rand.Intn(3000)

		req = req.WithContext(correlation.ContextWithMap(req.Context(), correlation.NewMap(correlation.MapUpdate{
			MultiKV: entries,
		})))

		ctx, span := tr.Start(
			trace.ContextWithRemoteSpanContext(req.Context(), spanCtx),
			"httpHandler",
			trace.WithAttributes(attrs...),
		)
		defer span.End()

		spanCtx = trace.SpanFromContext(ctx).SpanContext()
		logger := log.WithFields(log.Fields{
			"traceID": spanCtx.TraceID,
		})

		logger.Info("Handling request")

		span.AddEvent(ctx, "handling this...")

		time.Sleep(time.Duration(rnd) * time.Millisecond)
		statusCode := 0
		if rnd%2 == 0 {
			statusCode = recursiveCall(ctx)
		} else {
			if rnd%3 == 0 {
				statusCode = 500

			} else {
				statusCode = 200
			}
		}
		if statusCode == 500 {
			w.WriteHeader(http.StatusInternalServerError)
			trace.SpanFromContext(ctx).SetStatus(codes.Internal, "")
			logger.Error("Internal server error")

		} else {
			w.WriteHeader(http.StatusOK)
			trace.SpanFromContext(ctx).SetStatus(codes.OK, "")
		}
	})

	// Wrap our main handler, we pass empty handler ID so the middleware inferes
	// the handler label from the URL.
	h := std.Handler("", mdlw, mux)

	// Serve our handler.
	go func() {
		log.Printf("server listening at %s", srvAddr)
		if err := http.ListenAndServe(srvAddr, h); err != nil {
			log.Panicf("error while serving: %s", err)
		}
	}()

	promMux := http.NewServeMux()
	promMux.Handle("/metrics", promhttp.Handler())

	// Serve our metrics.
	go func() {
		log.Printf("metrics listening at %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, promMux); err != nil {
			log.Panicf("error while serving: %s", err)
		}
	}()

	// Wait until some signal is captured.
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
}

func recursiveCall(ctx context.Context) int {
	var body []byte
	client := http.DefaultClient
	tr := global.Tracer("recursiveCall")

	statusCode := 0
	err := tr.WithSpan(ctx, "client-call",
		func(ctx context.Context) error {
			spanCtx := trace.SpanFromContext(ctx).SpanContext()
			logCtx := log.WithFields(log.Fields{
				"traceID": spanCtx.TraceID,
			})

			logCtx.Info("Sending recursive request")

			req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

			ctx, req = httptrace.W3C(ctx, req)
			httptrace.Inject(ctx, req)

			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			body, err = ioutil.ReadAll(res.Body)
			_ = res.Body.Close()
			if res.StatusCode == 200 {
				trace.SpanFromContext(ctx).SetStatus(codes.OK, "Successful requres from client")
			} else {
				trace.SpanFromContext(ctx).SetStatus(codes.Internal, "Internal error from client")
			}
			statusCode = res.StatusCode

			return err
		})

	if err != nil {
		panic(err)
	}

	return statusCode

}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithAgentEndpoint(jaeger.CollectorEndpointFromEnv()),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "PingPong",
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		jaeger.WithDisabledFromEnv(),
		jaeger.WithProcessFromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		flush()
	}
}
