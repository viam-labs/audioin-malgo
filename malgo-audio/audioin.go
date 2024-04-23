package audioin

import (
	"context"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	malgo "github.com/gen2brain/malgo"
)

var Model = resource.NewModel("viam-labs", "sensor", "audio-in")

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
	sensor.logger.Debug("Current duration setting: ", sensor.defaultDuration*time.Millisecond)

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
		sensor.logger.Debug(message)
	})
	if err != nil {
		return err
	}

	infos, err := sensor.mContext.Devices(malgo.Capture)
	if err != nil {
		sensor.logger.Error("could not find capture devices")
		return err
	}

	for _, info := range infos {
		sensor.logger.Debug(info.Name(), info.IsDefault, info.String())
	}

	selectedDevice := infos[0]

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.DeviceID = selectedDevice.ID.Pointer()
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1

	sensor.pCapturedSampleCount = 0
	sensor.pCapturedSamples = make([]byte, 0)

	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	onRecvFrames := func(_outputSamples, inputSamples []byte, framecount uint32) {
		sampleCount := framecount * deviceConfig.Capture.Channels * sizeInBytes
		newCapturedSampleCount := sensor.pCapturedSampleCount + sampleCount
		sensor.pCapturedSamples = append(sensor.pCapturedSamples, inputSamples...)
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
	duration := sensor.defaultDuration

	if extra != nil {
		extraDuration, ok := extra["duration"].(float64)
		if ok {
			duration = time.Duration(extraDuration)
		}
	}
	sensor.logger.Debug("Gathering Readings...")

	sensor.pCapturedSampleCount = 0
	clear(sensor.pCapturedSamples)
	sensor.pCapturedSamples = make([]byte, 0)

	if err := sensor.mDevice.Start(); err != nil {
		sensor.logger.Error("Unable to start capture device")
		return nil, err
	}

	sensor.logger.Debug("Waiting for ", duration*time.Millisecond)
	time.Sleep(duration * time.Millisecond)

	if err := sensor.mDevice.Stop(); err != nil {
		sensor.logger.Error("Unable to stop capture device")
		return nil, err
	}

	sensor.logger.Debug("Sending Readings...")
	sensor.logger.Debug("Sample count: ", sensor.pCapturedSampleCount)
	readings := map[string]interface{}{
		"SampleCount": sensor.pCapturedSampleCount,
		"Samples":     sensor.pCapturedSamples,
	}

	return readings, nil
}

func (sensor *audioIn) Close(cts context.Context) error {
	sensor.mDevice.Uninit()

	if err := sensor.mContext.Uninit(); err != nil {
		return err
	}

	sensor.mContext.Free()
	return nil
}
