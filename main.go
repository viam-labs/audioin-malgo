package main

import (
	"context"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/viam-labs/audioin-malgo/malgo-audio"
)

func main() {
	utils.ContextualMain(mainWithArgs, logging.NewLogger("audioin-malgo"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) error {
	audioInModule, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	audioInModule.AddModelFromRegistry(ctx, sensor.API, audioin.Model)

	err = audioInModule.Start(ctx)
	defer audioInModule.Close(ctx)

	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
