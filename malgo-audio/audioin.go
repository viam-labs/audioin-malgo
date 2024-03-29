package audioin

import (
	"context"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/gen2brain/malgo"
)

var Model = resource.NewModel("viam", "sensor", "audio-in")

type Config struct {
	Duration time.Duration `json:"duration"`
}

func (cfg *Config) Validate(path string) ([]string, error) {
	return []string{}, nil
}

const defaultDuration time.Duration = 2000

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: func(
				ctx context.Context,
				_ resource.Dependencies,
				conf resource.Config,
				logger logging.Logger,
			) (sensor.Sensor, error) {
				return newAudioIn(ctx, conf, logger)
			},
		},
	)
}

func newAudioIn(ctx context.Context, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	audioSensor := audioIn{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
	}

	if err := audioSensor.Reconfigure(ctx, nil, conf); err != nil {
		return nil, err
	}

	return &audioSensor, nil
}

type audioIn struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable
	logger               logging.Logger
	defaultDuration      time.Duration
	mContext             *malgo.AllocatedContext
	mDevice              *malgo.Device
	pCapturedSamples     []byte
	pCapturedSampleCount uint32

	mu sync.RWMutex
}

func (sensor *audioIn) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	parsedConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}

	duration := parsedConf.Duration
	if duration == 0 {
		duration = defaultDuration
	}

	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	sensor.defaultDuration = duration

	if sensor.mContext != nil || sensor.mDevice != nil {
		return nil
	}

	if sensor.mContext != nil {
		if err := sensor.mContext.Uninit(); err != nil {
			return err
		}
		sensor.mContext.Free()
	}

	sensor.mContext, err = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		sensor.logger.Info(message)
	})
	if err != nil {
		return err
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1

	sensor.pCapturedSampleCount = 0
	sensor.pCapturedSamples = make([]byte, 0)

	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		sampleCount := framecount * deviceConfig.Capture.Channels * sizeInBytes
		newCapturedSampleCount := sensor.pCapturedSampleCount + sampleCount
		sensor.pCapturedSamples = append(sensor.pCapturedSamples, pSample...)
		sensor.pCapturedSampleCount = newCapturedSampleCount
	}
	capturedCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	if sensor.mDevice != nil {
		sensor.mDevice.Uninit()
	}

	sensor.mDevice, err = malgo.InitDevice(sensor.mContext.Context, deviceConfig, capturedCallbacks)
	if err != nil {
		return err
	}

	return nil
}

func (sensor *audioIn) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, resource.ErrDoUnimplemented
}

func (sensor *audioIn) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	// duration := extra["duration"].(time.Duration)

	// if duration == 0 {
	// 	duration = sensor.defaultDuration
	// }
	duration := sensor.defaultDuration
	sensor.logger.Info("Gathering Readings...")

	err := sensor.mDevice.Start()
	if err != nil {
		sensor.logger.Error("Unable to start capture device")
		return nil, err
	}

	time.Sleep(duration * time.Millisecond)

	err = sensor.mDevice.Stop()

	if err != nil {
		sensor.logger.Error("Unable to stop capture device")
		return nil, err
	}

	defer func() {
		sensor.logger.Info("Resetting values")
		sensor.pCapturedSampleCount = 0
		clear(sensor.pCapturedSamples)
	}()

	sensor.logger.Info("Sending Readings...")

	return map[string]interface{}{
		"sampleCount": sensor.pCapturedSampleCount,
		"samples":     sensor.pCapturedSamples,
	}, nil
}

func (sensor *audioIn) Close(cts context.Context) error {
	sensor.mDevice.Uninit()

	if err := sensor.mContext.Uninit(); err != nil {
		return err
	}

	sensor.mContext.Free()
	return nil
}
