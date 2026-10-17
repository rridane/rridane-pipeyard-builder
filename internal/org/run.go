package org

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rridane/golang-pipeyard/pi"
	"github.com/rridane/golang-pipeyard/x/inj/container"
	"github.com/rridane/golang-pipeyard/x/inj/forge"
	"github.com/rridane/golang-pipeyard/x/inj/otel"
	"github.com/rridane/golang-pipeyard/x/inj/sys"
	"github.com/rridane/golang-pipeyard/x/lay"
)

// Run wires one project's pipeline and runs the stage the command line names.
// project is "owner/repo", fixed at compile time: it is what the release
// stages talk to, whatever the environment says.
// Every binary is given the same dependencies; a stage receives only those it
// declares, so the GitHub client reaches nothing but the release stages.
//
// A binary carries no version: two builds of the same source have to be byte
// for byte equal, and its release tag is what names it.
func Run(name, project string, stages []pi.Stage) {
	loadEnv()

	endpoint, insecure := otlpTarget()
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	owner, repo, _ := strings.Cut(project, "/")

	tracer := &otel.TracerInjector{
		ServiceName:              name,
		OtelExporterOtlpEndpoint: endpoint,
		Insecure:                 insecure,
		Protocol:                 protocol,
	}
	meter := &otel.MeterInjector{
		ServiceName:              name,
		OtelExporterOtlpEndpoint: endpoint,
		Insecure:                 insecure,
		Protocol:                 protocol,
	}
	logger := &otel.TelemetryLoggerInjector{
		ServiceName:              name,
		EnableOtelExport:         endpoint != "",
		OtelExporterOtlpEndpoint: endpoint,
		OtelInsecure:             insecure,
		Protocol:                 protocol,
	}

	pipe := pi.NewPipeline(pi.PipelineConfig{
		Name:             name,
		ShortDescription: "The pipeline of " + project,
		LongDescription: "The pipeline of " + project + ", built from the Pipeyard catalogue.\n" +
			"Each command is one stage; the CI engine calls them in order.",

		Injectors: []any{
			logger,
			&sys.IdentityInjector{},
			&sys.ShellExecutorInjector{Timeout: 15 * time.Minute},
			&container.ContainerBuilderInjector{},
			&forge.GitHubReleaseInjector{
				Token: os.Getenv("GITHUB_TOKEN"),
				Owner: owner,
				Repo:  repo,
			},
		},

		Observer: &lay.StageObserver{
			Tracer: tracer.Inject(name),
			Meter:  meter.Inject(name),
			Logger: logger.Inject(""),
			Shutdown: []func(context.Context){
				tracer.Shutdown, meter.Shutdown, logger.Shutdown,
			},
		},
	})

	for _, stage := range stages {
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
