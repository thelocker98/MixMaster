package mixmaster

import (
	"bytes"

	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	hid "github.com/sstallion/go-hid"
	"github.com/tarm/serial"
	serialenum "go.bug.st/serial"
	"go.uber.org/zap"
)

var DataData string

type Device struct {
	HidDev    map[string]string
	SerialDev map[string]string
}

type DeviceData struct {
	Id     string
	Volume []float32
	Button []bool
	err    error
}
type arduinoMsg struct {
	Id       string `json:"id"`
	Slidders []int  `json:"s"`
	Buttons  []int  `json:"b"`
}
type idMsg struct {
	Id string `json:"id"`
}

func ListHIDDevices(logger *zap.SugaredLogger) []string {
	logger = logger.Named("ListHIDDevices")

	// Initialize the hid package.
	if err := hid.Init(); err != nil {
		logger.Error("Failed to initialize HID")
	}

	var paths []string

	logger.Debug("Enumerating All HID Devices")
	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		for _, path := range paths {
			if info.Path == path {
				return nil
			}
		}
		paths = append(paths, info.Path)
		return nil
	})

	logger.Debug("Found: ", paths)

	return paths
}

func ListSerialDevices(logger *zap.SugaredLogger) []string {
	logger = logger.Named("ListSerialDevices")

	logger.Debug("Enumerating all Serial Devices")
	ports, err := serialenum.GetPortsList()
	if err != nil {
		logger.Error("Error enumerating Serial Devices")
		return nil
	}

	logger.Debug("Found", ports)

	return ports
}

func GetDevice(logger *zap.SugaredLogger) (*Device, error) {
	logger = logger.Named("GetDevice")

	logger.Debug("Getting Devices")

	deviceList := Device{
		HidDev:    make(map[string]string),
		SerialDev: make(map[string]string),
	}

	hidDeviceList := ListHIDDevices(logger)
	for _, device := range hidDeviceList {
		logger.Debug("Opening device: ", device)
		d, err := hid.OpenPath(device)

		if err != nil {
			//logger.Error("Error Opening device: ", err)
			continue
		}

		// Send a I to the HID device to trigger a device id push
		data := make([]byte, 64)
		data[0] = 'I'
		_, err = d.Write(data)
		if err != nil {
			// Close connection and end this current loop
			//logger.Error("Error Writing to device: ", err)
			d.Close()
			continue
		}

		// Read data from the current HID device. If device does not respone it 100 milliseconds then timeout
		_, err = d.ReadWithTimeout(data, 100*time.Millisecond)
		if err != nil {
			// Close connection and end this current loop
			d.Close()
			continue
		}
		// Close any Open Devices
		d.Close()

		// Remove trailing zeros from byte array
		var cleanData []byte
		for _, b := range data {
			if b == 0 {
				break
			}
			cleanData = append(cleanData, b)
		}
		logger.Debug("Recieved Data: ", string(cleanData))
		deviceList.HidDev[parseDeviceID(cleanData).Id] = device
	}

	serialDeviceList := ListSerialDevices(logger)

	for _, port := range serialDeviceList {
		c := &serial.Config{
			Name:        port,
			Baud:        115200,
			ReadTimeout: time.Millisecond * 500,
		}

		p, err := serial.OpenPort(c)
		if err != nil {
			fmt.Println("error: ", err)
			continue
		}

		_, err = p.Write([]byte("H"))
		if err != nil {
			return nil, err
		}

		buff := make([]byte, 512)
		var all []byte

		for {
			n, err := p.Read(buff)
			if err != nil {
				break
			}
			if n > 0 {
				all = append(all, buff[:n]...)
				if bytes.Contains(all, []byte("}")) {
					break
				}
			}
		}

		var cleanData []byte
		for _, value := range all {
			cleanData = append(cleanData, value)
			if value == '}' {
				break
			}
		}
		p.Close()

		deviceList.SerialDev[parseDeviceData(cleanData, false).Id] = port
	}

	// No Device Found
	return &deviceList, nil
}

func ReadDeviceDataHID(d *hid.Device, cfg *DeviceConfig) (*DeviceData, error) {
	buff := make([]byte, 64) // automatically filled with zeros
	buff[0] = 'D'
	if _, err := d.Write(buff); err != nil {
		return nil, errors.New("error writing data to device")
	}

	// Buffers for read/write
	in := make([]byte, 64)
	var clean []byte
	deadline := time.Now().Add(1000 * time.Millisecond)
	for time.Now().Before(deadline) {
		// Read requested state.
		if _, err := d.Read(in); err != nil {
			return nil, errors.New("error reading data from device")
		}
		// Remove trailing zeros
		for _, b := range in {
			if b == 0 {
				break
			}
			clean = append(clean, b)
		}
		// Convert to string
		if clean[0] != '{' {
			clean = []byte{}
			return nil, errors.New("recieved currupt data")
		} else if clean[len(clean)-1] == '}' {
			break
		}
	}

	values := parseDeviceData(clean, cfg.SlidderInvert)

	return values, nil
}

func ReadDeviceDataSerial(p serialenum.Port, cfg *DeviceConfig) (*DeviceData, error) {
	var data []byte
	buf := make([]byte, 1)

	_, err := p.Write([]byte{'H'})
	if err != nil {
		return nil, err
	}

	for {
		n, err := p.Read(buf)
		if err != nil {
			return nil, err
		}

		if n > 0 {
			data = append(data, buf[0])
			if buf[0] == '}' {
				break
			}
		}
	}

	// parse data from the device
	values := parseDeviceData(data, cfg.SlidderInvert)
	return values, nil
}

func parseDeviceData(buf []byte, invertSliders bool) *DeviceData {
	var msg arduinoMsg
	err := json.Unmarshal(buf, &msg)
	if err != nil {
		return &DeviceData{
			Id:     "",
			Volume: nil,
			Button: nil,
			err:    err,
		}
	}

	// Convert to normalized float values (0–1)
	volumes := make([]float32, len(msg.Slidders))
	for i, v := range msg.Slidders {
		if invertSliders {
			v = 255 - v
		}
		volumes[i] = roundFloat32(float32(v)/255.0, 2) // since Arduino sends 0–255
	}

	buttons := make([]bool, len(msg.Buttons))
	for i, btn := range msg.Buttons {
		buttons[i] = (btn == 1)
	}

	return &DeviceData{
		Id:     msg.Id,
		Volume: volumes,
		Button: buttons,
		err:    nil,
	}
}

func parseDeviceID(buf []byte) *idMsg {
	var msg idMsg
	err := json.Unmarshal(buf, &msg)
	fmt.Println(string(msg.Id))
	if err != nil {
		fmt.Println("error", err)
		return &idMsg{
			Id: msg.Id,
		}
	}

	return &idMsg{
		Id: msg.Id,
	}
}

func JoinDeviceData(devices []*ParsedAudioData) *ParsedAudioData {
	var device ParsedAudioData
	device.PulseApps = make(map[string]float32)
	device.MpirsApps = make(map[string]MpirsApp)
	device.MasterOuputs = make(map[string]float32)

	for _, dev := range devices {
		for appName, val := range dev.PulseApps {
			if dat, ok := device.PulseApps[appName]; ok {
				device.PulseApps[appName] = (val + dat) / 2
			} else {
				device.PulseApps[appName] = val
			}
		}
		for appName, val := range dev.MasterOuputs {
			if dat, ok := device.MasterOuputs[appName]; ok {
				device.MasterOuputs[appName] = (val + dat) / 2
			} else {
				device.MasterOuputs[appName] = val
			}
		}
		for appName, val := range dev.MpirsApps {
			if dat, ok := device.MpirsApps[appName]; ok {
				if device.MpirsApps[appName].Back != true {
					dat.Back = val.Back
				}
				if device.MpirsApps[appName].PausePlay != true {
					dat.PausePlay = val.PausePlay
				}
				if device.MpirsApps[appName].Next != true {
					dat.Next = val.Next
				}

				device.MpirsApps[appName] = dat
			} else {
				device.MpirsApps[appName] = val
			}
		}
	}
	return &device
}

func GetArrayAt[T any](array []T, index int) (T, error) {
	var zero T
	if index < 0 || index >= len(array) {
		return zero, fmt.Errorf("index %d out of range [0,%d)", index, len(array))
	}
	return array[index], nil
}

func roundFloat32(f float32, decimals int) float32 {
	pow := math.Pow10(decimals)
	return float32(math.Round(float64(f)*pow) / pow)
}

func ScanForDevices(a fyne.App, cfg *Config, logger *zap.SugaredLogger, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, serialNumbers *map[string]string) {
	logger = logger.Named("devices")

	logger.Debug("Starting Device Scan")

	// Get a list of all devices pluged into computer
	pastNum := len(*devices)

	dev, _ := GetDevice(logger)

	for serialNum, _ := range dev.HidDev {
		logger.Debug("Devices found", SerialNumbers)
		if serialNum == "" {
			break
		}
		(*serialNumbers)[serialNum] = "Not Used"
		for name, device := range cfg.Devices {
			if device.SerialNumber == serialNum {
				(*serialNumbers)[serialNum] = name
				break
			}
		}
	}
	for serialNum, _ := range dev.SerialDev {
		if serialNum == "" {
			break
		}
		(*serialNumbers)[serialNum] = "Not Used"
		for name, device := range cfg.Devices {
			if device.SerialNumber == serialNum {
				(*serialNumbers)[serialNum] = name
				break
			}
		}
	}

	// Loop Through Devices in the config and see if they are connected to the computer
	for deviceName, device := range cfg.Devices {
		tempDevice, err := NewMixMaster(dev, device.SerialNumber)
		if err != nil {
			continue
		}
		(*devices)[deviceName] = tempDevice
	}
	deviceList.Set([]string{})
	connectedDevices.Set([]bool{})
	for name, _ := range cfg.Devices {
		deviceList.Append(name)
		_, ok := (*devices)[name]
		connectedDevices.Append(ok)
	}
	if cfg.App.Notifications && len(*devices) > pastNum {
		a.SendNotification(fyne.NewNotification("Mixmaster", "Discovered New Device"))
	}
}
