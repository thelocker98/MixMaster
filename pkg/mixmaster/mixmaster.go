package mixmaster

import (
	"errors"

	"github.com/sstallion/go-hid"
	"github.com/tarm/serial"
)

func NewMixMaster(cfg *Config, dev *Device) error {
	// Define Data Channel
	var deviceData *DeviceData
	// Hash of Volume and Button slice to detect change from device
	var volumeHash string
	var buttonHash string

	// Create Pulse Client
	client, err := CreatePulseClient("MixMaster")
	if err != nil {
		return errors.New("could not creating Pulse client")
	}

	// Set up mpris
	mpris, err := MprisInitialize()
	if err != nil {
		return errors.New("could not creating mpirs client")
	}

	// Create channel to receive data
	dataChan := make(chan *DeviceData)

	// Check if hid device is specified
	if dev.HidDev != nil {
		d, err := hid.OpenPath(*dev.HidDev)
		if err != nil {
			return errors.New("could not opening HID device")
		}

		go ReadDeviceDataHID(d, cfg, dataChan)

		//Check if serial device is specified
	} else if dev.SerialDev != nil {
		s, err := serial.OpenPort(dev.SerialDev)
		if err != nil {
			return errors.New("could not opening Serial device")
		}

		go ReadDeviceDataSerial(s, cfg, dataChan)

		// end if no device is found
	} else {
		return errors.New("no device found")
	}

	for {
		// Read from channel whenever new data arrives
		deviceData = <-dataChan

		// Get pulse audio sessions
		sessions, err := client.GetAudioSessions()
		if err != nil {
			// could not get pulse audio sessions
			continue
		}

		// Get mpris sessions
		players, err := mpris.ConnectToApps(sessions)
		if err != nil {
			// could not get app media controls
			continue
		}

		// Check if the hash of the current volume match the hash of the last volume.
		// If the hash matchs do not change volume
		if hash, _ := HashSlice(deviceData.Volume); hash != volumeHash {
			volumeHash, _ = HashSlice(deviceData.Volume)
			sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
			sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
		}

		// Check if the hash of the current media state match the hash of the last media state.
		// If the hash matchs do not change media state
		if hash, _ := HashSlice(deviceData.Button); hash != buttonHash {
			buttonHash, _ = HashSlice(deviceData.Button)
			players.PausePlay(cfg, deviceData)
		}
	}
}
