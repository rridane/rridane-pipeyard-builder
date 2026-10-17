// Command rr-ci is the pipeline every Python repository of this organisation
// runs. Each stage is one invocation; the CI engine keeps the sequencing.
//
// A repository holds none of this. Its workflow downloads a pinned release and
// calls the stages; what a stage reads about the project comes from the
// project's own configuration, pyproject.toml and uv.lock.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/inj/container"
	"github.com/rridane/golang-pipeyard/x/inj/otel"
	"github.com/rridane/golang-pipeyard/x/inj/sys"
	"github.com/rridane/golang-pipeyard/x/lay"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

const serviceName = "rr-ci"

func main() {
	loadEnv()

	endpoint, insecure := otlpTarget()
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	tracer := &otel.TracerInjector{
		ServiceName:              serviceName,
		OtelExporterOtlpEndpoint: endpoint,
		Insecure:                 insecure,
		Protocol:                 protocol,
	}
	meter := &otel.MeterInjector{
		ServiceName:              serviceName,
		OtelExporterOtlpEndpoint: endpoint,
		Insecure:                 insecure,
		Protocol:                 protocol,
	}
	logger := &otel.TelemetryLoggerInjector{
		ServiceName:              serviceName,
		EnableOtelExport:         endpoint != "",
		OtelExporterOtlpEndpoint: endpoint,
		OtelInsecure:             insecure,
		Protocol:                 protocol,
	}

	pipe := pi.NewPipeline(pi.PipelineConfig{
		Name:             serviceName,
		Version:          version,
		ShortDescription: "The pipeline of rridane's Python repositories",
		LongDescription: "Holds a repository to the organisation's rules, looks for secrets in its\n" +
			"history, checks the Python and audits the locked dependencies.\n" +
			"Each command is one stage; the CI engine calls them in order.",

		Injectors: []any{
			logger,
			&sys.IdentityInjector{},
			&sys.ShellExecutorInjector{Timeout: 15 * time.Minute},
			&container.ContainerBuilderInjector{},
		},

		Observer: &lay.StageObserver{
			Tracer: tracer.Inject(serviceName),
			Meter:  meter.Inject(serviceName),
			Logger: logger.Inject(""),
			Shutdown: []func(context.Context){
				tracer.Shutdown, meter.Shutdown, logger.Shutdown,
			},
		},
	})

	for _, stage := range stages() {
		pipe.RegisterStage(stage)
	}
	pipe.Run(context.Background())
}

// loadEnv reads a .env when there is one, so a laptop run picks up the same
// settings the CI gets from its own environment. A missing file is normal.
func loadEnv() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: .env not loaded: %v\n", err)
	}
}

// otlpTarget reads the standard variable, which is a URL, and hands back the
// host:port the gRPC exporters take plus the transport the scheme asked for.
// Unset, the exporters fall back to noop and nothing is recorded.
func otlpTarget() (endpoint string, insecure bool) {
	raw := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	insecure = strings.HasPrefix(raw, "http://")

	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimPrefix(raw, "https://")

	return strings.TrimSuffix(raw, "/"), insecure
}
