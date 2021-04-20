package main

import (
	"errors"
	"log"
	"net"
	"strings"
)

func getIPbyName(name string) (string, error) {
	ifaces, err := net.Interfaces() // Interfaces returns a list of the system's network interfaces.
	if err == nil {
		// handle err
		for _, i := range ifaces { // go through each interface associated with this system
			log.Println(i)
			if name == i.Name {
				log.Println("FOUND")
				if addrs, err := i.Addrs(); err == nil { // get slice of addresses associated with this interface
					// handle err
					for _, addr := range addrs { // go through each address associated with this interface
						log.Println("addr type:", addr.Network())
						log.Println("addr type:", addr.String())
						if !strings.Contains(addr.String(), ":") {
							return strings.Split(addr.String(), "/")[0], nil
						}
					}
				} else {
					return "", errors.New("getIPbyName(): " + err.Error())
				}
			}
		}
		return "", errors.New("getIPbyName(): interface not found")
	}
	return "", errors.New("getIPbyName(): " + err.Error())
}
