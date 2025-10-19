package mixmaster

import (
	"errors"
	"fmt"

	"github.com/jfreymuth/pulse/proto"
)

type pulseClient struct {
	client *proto.Client
}

type PulseSessions struct {
	Masters map[string]*proto.GetSinkInfoReply
	Apps    map[string][]*proto.GetSinkInputInfoReply
	//AppNames []string
}

const maxVolume = 0x10000

func CreatePulseClient(applicationName string) (*pulseClient, error) {
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
	return &pulseClient{client: client}, nil
}

func (c *pulseClient) GetPulseSessions() (PulseSessions, error) {
	data := PulseSessions{
		Masters: make(map[string]*proto.GetSinkInfoReply),
		Apps:    make(map[string][]*proto.GetSinkInputInfoReply),
	}

	requestApp := proto.GetSinkInputInfoList{}
	replyApp := proto.GetSinkInputInfoListReply{}

	if err := c.client.Request(&requestApp, &replyApp); err != nil {
		return PulseSessions{}, err
	}

	for _, info := range replyApp {
		// Check for application process binary
		appName, okApp := info.Properties["application.process.binary"]
		// Check for second name
		nodeName, okNode := info.Properties["node.name"]

		if okApp {
			data.Apps[appName.String()] = append(data.Apps[appName.String()], info)
			//data.AppNames = append(data.AppNames, appName.String())
		} else if okNode {
			data.Apps[nodeName.String()] = append(data.Apps[nodeName.String()], info)
			//data.AppNames = append(data.AppNames, nodeName.String())
		} else {
			continue
		}
	}

	requestMaster := proto.GetSinkInfoList{}
	replyMaster := proto.GetSinkInfoListReply{}

	if err := c.client.Request(&requestMaster, &replyMaster); err != nil {
		return PulseSessions{}, err
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

func (sessions *PulseSessions) ChangeAppVolume(parsedData PulseApps, c *pulseClient) error {
	for appName, vol := range parsedData {
		if appName != "unmapped" {
			pulseSink, ok := sessions.Apps[appName]
			if ok {
				for _, indices := range pulseSink {
					request := proto.SetSinkInputVolume{
						SinkInputIndex: indices.SinkInputIndex,
						ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
					}
					if err := c.client.Request(&request, nil); err != nil {
						fmt.Println("error:", err)
						continue
					}
				}
			}
		} else {
			for appName, pulseSink := range sessions.Apps {
				if _, ok := parsedData[appName]; !ok {
					for _, indices := range pulseSink {
						request := proto.SetSinkInputVolume{
							SinkInputIndex: indices.SinkInputIndex,
							ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
						}
						if err := c.client.Request(&request, nil); err != nil {
							fmt.Println("error:", err)
							continue
						}
					}
				}
			}
		}
	}
	return nil
}

func (sessions *PulseSessions) ChangeMasterVolume(parsedData PulseApps, c *pulseClient) error {
	for pulseOutputName, vol := range parsedData {
		if pulseOutputName == "master" {
			for outputName, pulseSink := range sessions.Masters {
				if _, ok := parsedData[outputName]; !ok {
					request := proto.SetSinkVolume{
						SinkIndex:      pulseSink.SinkIndex,
						ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
					}
					if err := c.client.Request(&request, nil); err != nil {
						fmt.Println("error:", err)
						continue
					}
				}
			}
		} else {
			pulseSink, ok := sessions.Masters[pulseOutputName]
			if ok {
				request := proto.SetSinkVolume{
					SinkIndex:      pulseSink.SinkIndex,
					ChannelVolumes: []uint32{uint32(maxVolume * vol)}, // volume
				}
				if err := c.client.Request(&request, nil); err != nil {
					fmt.Println("error:", err)
					continue
				}
			}
		}
	}
	return nil
}

func (c *pulseClient) GetMasterVolume(name string) (*float32, error) {
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

func (c *pulseClient) GetAppVolume(appName string) (*float32, error) {
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

func (sessions *PulseSessions) DisplayAppNames() {
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

func parseChannelVolumes(volumes []uint32) *float32 {
	var level uint32

	for _, volume := range volumes {
		level += volume
	}

	vol := float32(level) / float32(len(volumes)) / float32(maxVolume)

	return &vol
}
