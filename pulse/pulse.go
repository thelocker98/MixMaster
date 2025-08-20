package pulse

import (
	"errors"
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"github.com/jfreymuth/pulse/proto"
)

type pulseAudio struct {
	client *proto.Client
}

type appSessions struct {
	Masters map[string]*proto.GetSinkInfoReply
	Apps    map[string]*proto.GetSinkInputInfoReply
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

func (sessions *appSessions) ChangeAppVolume(cfg *config.Config, volume []float32, c *pulseAudio) error {
	unmappedSlider, unmappedOk := cfg.AppSliderMapping["unmapped"]

	for name, info := range sessions.Apps {
		val, Ok := cfg.AppSliderMapping[name]

		if Ok {
			request := proto.SetSinkInputVolume{
				SinkInputIndex: info.SinkInputIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * volume[val])}, // volume
			}

			if err := c.client.Request(&request, nil); err != nil {
				continue
			}
		} else if unmappedOk {
			request := proto.SetSinkInputVolume{
				SinkInputIndex: info.SinkInputIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * volume[unmappedSlider])}, // volume
			}

			if err := c.client.Request(&request, nil); err != nil {
				continue
			}
		}

	}
	return nil
}

func (sessions *appSessions) ChangeMasterVolume(cfg *config.Config, volume []float32, c *pulseAudio) error {
	masterSlider, masterOk := cfg.MasterSliderMapping["master"]

	for name, info := range sessions.Masters {
		val, Ok := cfg.MasterSliderMapping[name]

		if masterOk {
			val = masterSlider
		}

		if Ok || masterOk {
			request := proto.SetSinkVolume{
				SinkIndex:      info.SinkIndex,
				ChannelVolumes: []uint32{uint32(maxVolume * volume[val])}, // volume
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

func (c *pulseAudio) GetAudioSessions() (*appSessions, error) {
	data := &appSessions{
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
		} else if okNode {
			data.Apps[nodeName.String()] = info
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

func (sessions *appSessions) DisplayAppNames() {
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
