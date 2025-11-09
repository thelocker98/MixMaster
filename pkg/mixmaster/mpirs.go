package mixmaster

import (
	"errors"
	"strings"

	"github.com/Endg4meZer0/go-mpris"
	"github.com/godbus/dbus/v5"
)

type mpirsClient struct {
	Conn *dbus.Conn
}

type MpirsSessions map[string]*mpris.Player

func MprisInitialize() (*mpirsClient, error) {
	conn, err := dbus.SessionBus()

	if err != nil {
		return nil, err
	}

	return &mpirsClient{Conn: conn}, nil
}

func (conn *mpirsClient) GetMpirsSessions() (MpirsSessions, error) {
	audioPlayers := make(MpirsSessions)

	appNames, err := mpris.List(conn.Conn)
	if err != nil {
		return nil, errors.New("Error getting mpirs app list")
	}
	for _, appName := range appNames {
		player := mpris.New(conn.Conn, appName)
		playerName := strings.ToLower(strings.Split(player.GetShortName(), ".")[0])
		audioPlayers[playerName] = player
	}
	return audioPlayers, nil
}

func (sessions *MpirsSessions) MediaControls(parsedData MpirsApps, c *mpirsClient) error {
	for appName, appData := range parsedData {
		mpirsSocket, ok := (*sessions)[appName]
		if ok {
			if appData.Back {
				mpirsSocket.Previous()
			}
			if appData.PausePlay {
				mpirsSocket.PlayPause()
			}
			if appData.Next {
				mpirsSocket.Next()
			}
		}

	}
	return nil
}
