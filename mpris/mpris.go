package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Endg4meZer0/go-mpris"
	"github.com/godbus/dbus/v5"
)

func mpris_test() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	names, err := mpris.List(conn)
	if err != nil {
		panic(err)
	}
	if len(names) == 0 {
		log.Fatal("No players found")
	}

	for _, name := range names {
		player := mpris.New(conn, name)

		status, err := player.GetVolume()
		if err != nil {
			log.Fatalf("Could not get current playback status: %s", err)
		}

		identity, err := player.GetIdentity()
		if err != nil {
			log.Fatalf("Could not get current playback status: %s", err)
		}
		meta, err := player.GetMetadata()
		if err != nil {
			log.Fatalf("Could not get current playback status: %s", err)
		}

		fmt.Println(status)
		fmt.Println(identity)
		fmt.Println(meta)

		player.SetVolume(100)
		time.Sleep(5 * time.Second)
		player.SetVolume(-100)

	}

}
