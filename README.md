# audioin-malgo

**Currently unpublished and a proof of concept**

This module implements the [`rdk:component:sensor` API](https://docs.viam.com/components/sensor/) in an `viam-labs:sensor:audio-in` model.
With this model, you can capture audio data from any audio input, like a microphone, as a sensor.

Uses [malgo](https://github.com/gen2brain/malgo) as the cross-platform audio driver.

> [!NOTE]
> For more information, see [Modular Resources](https://docs.viam.com/registry/#modular-resources).

## Requirements

For Linux, make sure [PulseAudio](https://en.wikipedia.org/wiki/PulseAudio), [PipeWire](https://en.wikipedia.org/wiki/PipeWire), or [ALSA](https://www.maketecheasier.com/alsa-utilities-manage-linux-audio-command-line/) is available and set up for the audio input device.

For MacOS, no setup is required.

## Building

The `make build` command is used to compile the cross-platform binaries to distribute as the entrypoint of the module. It requires [`zig`](https://ziglang.org/) as the compiler. The `.mise.toml` config will automatically install `zig` using [`mise`](https://mise.jdx.dev/).
