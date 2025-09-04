package audio

import (
	"errors"
	"fmt"
	"strings"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pulse"
	"gitea.locker98.com/locker98/Mixmaster/serial"
	"github.com/Endg4meZer0/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/sahilm/fuzzy"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type DBusConnection struct {
	Conn *dbus.Conn
}

type Players map[string]*mpris.Player

func MprisInitialize() (*DBusConnection, error) {
	conn, err := dbus.SessionBus()

	if err != nil {
		return nil, err
	}

	return &DBusConnection{Conn: conn}, nil
}

func (conn *DBusConnection) ConnectToApps(sessions *pulse.AppSessions) (*Players, error) {
	audioPlayers := make(Players)

	names, err := mpris.List(conn.Conn)

	if err != nil {
		return nil, err
	}

	for _, name := range names {
		player := mpris.New(conn.Conn, name)
		playerName := strings.ToLower(strings.Split(player.GetShortName(), ".")[0])

		audioPlayers[playerName] = player
	}
	return &audioPlayers, nil
}

func (p Players) PausePlay(cfg *config.Config, deviceData *serial.DeviceData) {
	for name, val := range p {
		if _, ok := cfg.Buttons[name]; !ok {
			continue
		}
		if deviceData.Button[cfg.Buttons[name]] {
			val.PlayPause()
		}
	}
}

// BestMatch finds the closest match in candidates for the given input.
func bestMatch(input string, candidates []string) (*string, error) {
	best := ""
	bestDistance := -1

	for _, candidate := range candidates {
		dist := levenshtein.DistanceForStrings([]rune(input), []rune(candidate), levenshtein.DefaultOptions)

		if bestDistance == -1 || dist < bestDistance {
			best = candidate
			bestDistance = dist
		}
	}

	if bestDistance > 10 {
		return nil, fmt.Errorf("No match found. Distance was: %d", bestDistance)
	}
	return &best, nil
}

// BestMatch finds the best fuzzy match for input among candidates.
func bestMatchFZ(input string, candidates []string) (*string, error) {
	if len(candidates) == 0 {
		return nil, errors.New("No input")
	}

	// fuzzy.Find returns matches sorted by score (best first)
	matches := fuzzy.Find(input, candidates)

	if len(matches) == 0 {
		return nil, errors.New("No items to be matched to")
	}

	// First match is the best one
	best := matches[0]
	return &candidates[best.Index], nil
}
