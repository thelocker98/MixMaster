package pulse

import (
	"errors"
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/device"
	"github.com/jfreymuth/pulse/proto"
)

type pulseAudio struct {
	client *proto.Client
}

type AppSessions struct {
	Masters  map[string]*proto.GetSinkInfoReply
	Apps     map[string]*proto.GetSinkInputInfoReply
	AppNames []string
}

const maxVolume = 0x10000

func (c *pulseAudio) GetAppVolume(appName string) (*float32, error) {
	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := c.client.Request(&request, &reply); err != nil {
		return nil, err
	}

	for _, info := range reply {
		name, ok := info.Properties["application.process.binary"]
		if !ok {
			return nil, errors.New("Failed to find app name")
		}

		if name.String() == appName {
			return parseChannelVolumes(info.ChannelVolumes), nil
		}
	}
	return nil, fmt.Errorf(": %s", appName)
}

func (sessions *AppSessions) ChangeAppVolume(cfg *config.Config, volume []float32, c *pulseAudio) error {
	unmappedSlider, unmappedOk := cfg.AppSlidderMapping["unmapped"]

	for name, info := range sessions.Apps {
		val, Ok := cfg.AppSlidderMapping[name]

		if Ok && val.Slidder != -1 {
			vol, err := device.GetAt(volume, val.Slidder)

			if err != nil {
				continue
			}

			request := proto.SetSinkInputVolume{
				SinkInputIndex: info.SinkInputIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
			}
			if err := c.client.Request(&request, nil); err != nil {
				continue
			}
		} else if unmappedOk {
			vol, err := device.GetAt(volume, unmappedSlider.Slidder)

			if err != nil {
				continue
			}

			request := proto.SetSinkInputVolume{
				SinkInputIndex: info.SinkInputIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
			}

			if err := c.client.Request(&request, nil); err != nil {
				continue
			}
		}

	}
	return nil
}

func (sessions *AppSessions) ChangeMasterVolume(cfg *config.Config, volume []float32, c *pulseAudio) error {
	masterSlider, masterOk := cfg.MasterSlidderMapping["master"]
	masterSlider--

	for name, info := range sessions.Masters {
		pin, Ok := cfg.MasterSlidderMapping[name]
		// Minus 1
		pin--
		if masterOk {
			pin = masterSlider
		}

		// Get volume
		vol, err := device.GetAt(volume, pin)
		if err != nil {
			continue
		}

		if Ok || masterOk {
			request := proto.SetSinkVolume{
				SinkIndex:      info.SinkIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
			}

			if err := c.client.Request(&request, nil); err != nil {
				continue
			}
		}
	}
	return nil
}

func (c *pulseAudio) GetMasterVolume(name string) (*float32, error) {
	request := proto.GetSinkInfoList{}
	reply := proto.GetSinkInfoListReply{}
	totalvolume := []float32{}

	if err := c.client.Request(&request, &reply); err != nil {
		return nil, err
	}

	for _, info := range reply {
		nodeName, ok := info.Properties["device.profile.name"]
		if !ok {
			return nil, errors.New("Error finding audio output name")
		}

		if nodeName.String() == name || name == "master" {
			level := parseChannelVolumes(info.ChannelVolumes)
			if name != "master" {
				return level, nil
			} else {
				totalvolume = append(totalvolume, *level)
			}
		}
	}

	if name == "master" {
		var total float32
		for _, val := range totalvolume {
			total += val
		}
		result := total / float32(len(totalvolume))
		return &result, nil
	}
	return nil, fmt.Errorf("Failed to find audio ouput named: %s", name)
}

func (c *pulseAudio) GetAudioSessions() (*AppSessions, error) {
	data := &AppSessions{
		Masters: make(map[string]*proto.GetSinkInfoReply),
		Apps:    make(map[string]*proto.GetSinkInputInfoReply),
	}

	requestApp := proto.GetSinkInputInfoList{}
	replyApp := proto.GetSinkInputInfoListReply{}

	if err := c.client.Request(&requestApp, &replyApp); err != nil {
		return nil, err
	}

	for _, info := range replyApp {
		appName, okApp := info.Properties["application.process.binary"]
		nodeName, okNode := info.Properties["node.name"]

		if okApp {
			data.Apps[appName.String()] = info
			data.AppNames = append(data.AppNames, appName.String())
		} else if okNode {
			data.Apps[nodeName.String()] = info
			data.AppNames = append(data.AppNames, appName.String())
		} else {
			continue
		}
	}

	requestMaster := proto.GetSinkInfoList{}
	replyMaster := proto.GetSinkInfoListReply{}

	if err := c.client.Request(&requestMaster, &replyMaster); err != nil {
		return nil, err
	}

	for _, info := range replyMaster {
		nodeName, ok := info.Properties["device.profile.name"]
		if !ok {
			continue
		}
		data.Masters[nodeName.String()] = info

	}

	return data, nil
}

func CreatePulseClient(applicationName string) (*pulseAudio, error) {
	client, _, err := proto.Connect("")
	if err != nil {
		return nil, err
	}

	request := proto.SetClientName{
		Props: proto.PropList{
			"application.name": proto.PropListString(applicationName),
		},
	}
	reply := proto.SetClientNameReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, err
	}
	return &pulseAudio{client: client}, nil
}

func parseChannelVolumes(volumes []uint32) *float32 {
	var level uint32

	for _, volume := range volumes {
		level += volume
	}

	vol := float32(level) / float32(len(volumes)) / float32(maxVolume)

	return &vol
}

func (sessions *AppSessions) DisplayAppNames() {
	fmt.Println("       Audio Devices and Apps:")
	fmt.Println("List of Open Audio Apps:")
	fmt.Println("- unmapped")
	for name, _ := range sessions.Apps {
		fmt.Printf("- %s\n", name)
	}

	fmt.Println("\nList of Audio Output Devices:")
	fmt.Println("- master")
	for output, _ := range sessions.Masters {
		fmt.Printf("- %s\n", output)
	}
	fmt.Println("")
}
