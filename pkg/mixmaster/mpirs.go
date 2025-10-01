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

func (conn *mpirsClient) GetMpirsSessions() (*MpirsSessions, error) {
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
	return &audioPlayers, nil
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

// // Unused Code
// func (conn *DBusConnection) ConnectToApps(sessions *PulseSessions) (*Players, error) {
// 	audioPlayers := make(Players)

// 	names, err := mpris.List(conn.Conn)

// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, name := range names {
// 		player := mpris.New(conn.Conn, name)
// 		playerName := strings.ToLower(strings.Split(player.GetShortName(), ".")[0])

// 		audioPlayers[playerName] = player
// 	}
// 	return &audioPlayers, nil
// }

// // BestMatch finds the closest match in candidates for the given input.
// func bestMatch(input string, candidates []string) (*string, error) {
// 	best := ""
// 	bestDistance := -1

// 	for _, candidate := range candidates {
// 		dist := levenshtein.DistanceForStrings([]rune(input), []rune(candidate), levenshtein.DefaultOptions)

// 		if bestDistance == -1 || dist < bestDistance {
// 			best = candidate
// 			bestDistance = dist
// 		}
// 	}

// 	if bestDistance > 10 {
// 		return nil, fmt.Errorf("No match found. Distance was: %d", bestDistance)
// 	}
// 	return &best, nil
// }

// // BestMatch finds the best fuzzy match for input among candidates.
// func bestMatchFZ(input string, candidates []string) (*string, error) {
// 	if len(candidates) == 0 {
// 		return nil, errors.New("No input")
// 	}

// 	// fuzzy.Find returns matches sorted by score (best first)
// 	matches := fuzzy.Find(input, candidates)

// 	if len(matches) == 0 {
// 		return nil, errors.New("No items to be matched to")
// 	}

// 	// First match is the best one
// 	best := matches[0]
// 	return &candidates[best.Index], nil
// }
