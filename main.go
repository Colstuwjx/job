package main

import (
	"flag"

	"github.com/Colstuwjx/job/config"
	"github.com/Colstuwjx/job/env"
	"github.com/Colstuwjx/job/impl"
	ilogger "github.com/Colstuwjx/job/impl/logger"
	"github.com/Colstuwjx/job/logger"
	"github.com/Colstuwjx/job/runtime"
	"github.com/Colstuwjx/job/utils"
)

func main() {
	// Get parameters
	configPath := flag.String("c", "", "Specify the yaml config file path")
	flag.Parse()

	// Missing config file
	if configPath == nil || utils.IsEmptyStr(*configPath) {
		flag.Usage()
		logger.Fatal("Config file should be specified")
	}

	// Load configurations
	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		logger.Fatalf("Failed to load configurations with error: %s\n", err)
	}

	// Set job context initializer
	runtime.JobService.SetJobContextInitializer(func(ctx *env.Context) (env.JobContext, error) {
		jobCtx := impl.NewContext(ctx.SystemContext)
		if err := jobCtx.Init(); err != nil {
			return nil, err
		}

		return jobCtx, nil
	})

	// New logger for job service
	sLogger := ilogger.NewServiceLogger(config.GetLogLevel())
	logger.SetLogger(sLogger)

	// Register jobs first
	runtime.Register(impl.KnownJobDemo, (*impl.DemoJob)(nil))

	// Start
	runtime.JobService.LoadAndRun()
}
