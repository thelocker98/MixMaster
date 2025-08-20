
# MixMaster

**MixMaster** is a hardware + software tool that lets you control the audio and playback of different applications on your Linux computer using a microcontroller. Adjust volumes, play, skip, and manage multiple apps—all from a single physical device.

> ⚠️ **Work in Progress**: This project is currently under development. Features may change, and stability is not guaranteed.

## Features
- Control the volume of individual applications in real-time
- Play, pause, and skip tracks across supported apps
- Works through a microcontroller plugged into your computer
- Designed for Linux systems (currently does **not** support Windows or macOS)

## Requirements
- Linux OS
- Go 1.20+
- Microcontroller (e.g., Arduino) connected via USB/COM port

## Usage
1. Connect your microcontroller to your Linux computer
2. Configure the apps and sliders in the provided YAML config file
3. Run MixMaster to start controlling your system audio

> Example config snippet:

```yaml
slider_mapping:
  master: 0
  chrome: 1
  spotify: 2

invert_sliders: false
com_port: /dev/ttyUSB0
baud_rate: 9600
noise_reduction: default
```

## Status
- [x] Basic volume control
- [x] App mapping via config file
- [ ] Playback controls (play, pause, skip)
- [ ] Arduino Code

## Contributing
Contributions are welcome! If you want to help develop MixMaster, feel free to submit issues or pull requests.
