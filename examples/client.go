package main

import (
	"context"
	"encoding/base64"
	"os"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const (
	address    = "my-machine.viam.cloud" /* TODO: fill in correct address */
	api_key_id = ""                      /* TODO: fill in API Key ID */
	api_key    = ""                      /* TODO: fill in API Key */
)

func main() {
	logger := logging.NewDebugLogger("client")
	machine, err := client.New(
		context.Background(),
		address,
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			/* Replace "<API-KEY-ID>" (including brackets) with your machine's api key id */
			api_key_id,
			rpc.Credentials{
				Type: rpc.CredentialsTypeAPIKey,
				/* Replace "<API-KEY>" (including brackets) with your machine's api key */
				Payload: api_key,
			})),
	)
	if err != nil {
		logger.Fatal(err)
		return
	}

	defer machine.Close(context.Background())
	logger.Info("Resources:")
	logger.Info(machine.ResourceNames())

	audioIn, err := sensor.FromRobot(machine, "audioin")
	if err != nil {
		logger.Fatal(err)
		return
	}

	readings, err := audioIn.Readings(context.Background(), map[string]interface{}{
		"duration": 500,
	})
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("Readings: ", readings["Samples"])
	audioData, err := base64.StdEncoding.DecodeString(readings["Samples"].(string))
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("Decoded bytes: ", audioData)

	file, err := os.Create("test.wav")
	if err != nil {
		logger.Fatal(err)
		return
	}
	defer file.Close()

	// convert captured audio data into format required by WAV encoder
	buf := &audio.IntBuffer{
		Data: make([]int, len(audioData)/2),
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  44100,
		},
		SourceBitDepth: 16,
	}
	for i := 0; i < len(audioData); i += 2 {
		buf.Data[i/2] = int(int16(audioData[i]) | int16(audioData[i+1])<<8)
	}

	// save encoded audio to file
	encoder := wav.NewEncoder(file, buf.Format.SampleRate, buf.SourceBitDepth, buf.Format.NumChannels, 1)

	if err := encoder.Write(buf); err != nil {
		logger.Fatal(err)
		return
	}

	if err := encoder.Close(); err != nil {
		logger.Fatal(err)
		return
	}

	logger.Info("Saved test.wav successfully")
}
