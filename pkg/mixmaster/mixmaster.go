package mixmaster

import (
	"errors"

	"github.com/sstallion/go-hid"

	"go.bug.st/serial"
)

type MixMasterInstance struct {
	HID    *hid.Device
	Serial serial.Port
}

func NewMixMaster(dev *Device, serialNum int64) (*MixMasterInstance, error) {
	// Check if hid device is specified
	if devicePath, ok := dev.HidDev[serialNum]; ok {
		d, err := hid.OpenPath(devicePath)
		if err != nil {
			return nil, errors.New("could not opening HID device")
		}
		return &MixMasterInstance{HID: d}, nil

		//Check if serial device is specified
	} else if devicePath, ok := dev.SerialDev[serialNum]; ok {
		s, err := serial.Open(devicePath, &serial.Mode{BaudRate: 115200})
		if err != nil {
			return nil, errors.New("could not opening Serial device")
		}

		return &MixMasterInstance{Serial: s}, nil

	} else {
		return nil, errors.New("no device found")
	}
}

func (dev MixMasterInstance) Pull(cfg *DeviceConfig) (*ParsedAudioData, error) {
	if dev.HID != nil {
		deviceData, err := ReadDeviceDataHID(dev.HID, cfg)
		if err != nil {
			return nil, err
		}
		data, err := deviceData.parseDataConfig(cfg)
		return &data, err

	} else if dev.Serial != nil {
		deviceData, err := ReadDeviceDataSerial(dev.Serial, cfg)
		if err != nil {
			return nil, err
		}
		data, err := deviceData.parseDataConfig(cfg)
		return &data, err

	} else {
		return nil, errors.New("No Device")
	}
}

// // Get pulse audio sessions
// sessions, err := client.GetAudioSessions()
// if err != nil {
// 	// could not get pulse audio sessions
// 	continue
// }

// // Get mpris sessions
// players, err := mpris.ConnectToApps(sessions)
// if err != nil {
// 	// could not get app media controls
// 	continue
// }

// // Check if the hash of the current volume match the hash of the last volume.
// // If the hash matchs do not change volume
// if hash, _ := HashSlice(deviceData.Volume); hash != volumeHash {
// 	volumeHash, _ = HashSlice(deviceData.Volume)
// 	sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
// 	sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
// }

// // Check if the hash of the current media state match the hash of the last media state.
// // If the hash matchs do not change media state
// if hash, _ := HashSlice(deviceData.Button); hash != buttonHash {
// 	buttonHash, _ = HashSlice(deviceData.Button)
// 	players.PausePlay(cfg, deviceData)
// }
