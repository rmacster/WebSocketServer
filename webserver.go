package main

/*
	SEE ALSO:
		Not used here but seems to be an AMAZING go framework:
			https://github.com/gofiber/fiber

	Primarily from:
		https://gowebexamples.com/websockets/
	And:
		https://github.com/denji/golang-tls  (Excellent resource!)

	Generate certs in ~/.ssh with the command:
		openssl req -x509 -newkey rsa:4096 -keyout ~/.ssh/key.pem -out ~/.ssh/cert.pem -days 365 -nodes

	Uses two env vars:
		UseInsecureCerts:	If set to "true", will set tlsConfig.InsecureSkipVerify to true.
		CertsPath:			Path to certs, including trailing slash.
	If you have to run as sudo, use the command "sudo -E ./webserver" to ensure sudo has access to your env vars.
*/

import (
	"log"
	"net/http"
	"os"

	"crypto/tls"

	"github.com/gorilla/websocket"
)

var send = make(chan string)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func wsStart() {
	if addr, err := getIPbyName("eth0"); err == nil {
		log.Println("wsStart(): eth0 IP: " + addr)
	} else {
		log.Println("wsStart(): " + err.Error())
	}
	go func() {
		// this is a standalone http server that only handles redirect to https
		go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		}))

		// beginning of https server

		// for better efficiency and clarity, handle routes through a mux
		mux := http.NewServeMux()

		// Standard HTML page
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "index.html")
		})

		// this page has a websocket client that connects via /ws.
		mux.HandleFunc("/websockets", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "websockets.html")
		})

		// upgrade to websocket connection
		mux.HandleFunc("/ws", wsComms)

		tlsConfig := &tls.Config{ // See: https://github.com/denji/golang-tls
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		// set an env var to use self-signed certs
		//if strings.ToLower(os.Getenv("UseInsecureCerts")) == "true" {
		tlsConfig.InsecureSkipVerify = true
		//}

		srv := &http.Server{
			Addr:         ":443",
			Handler:      mux,
			TLSConfig:    tlsConfig,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0), // enable HTTP/2
		}
		log.Println(os.Getenv("CertsPath"))
		log.Println(os.Getenv("CertsPath"))
		log.Fatal(srv.ListenAndServeTLS(os.Getenv("CertsPath")+"cert.pem", os.Getenv("CertsPath")+"key.pem"))
	}()
}

// wsComms - Handles communications to/from websocket
func wsComms(w http.ResponseWriter, r *http.Request) {
	if ws, err := upgrader.Upgrade(w, r, nil); err == nil { // error ignored for sake of simplicity
		defer ws.Close()

		log.Printf("wsComms(): Connection from %s upgraded...\n", ws.RemoteAddr())

		// handle comms TO the websocket
		go func(s *websocket.Conn) {
			for {
				select {
				case m := <-send:
					// In actual practice, we'd probably use ws.WriteJSON instead.
					if err := s.WriteMessage(websocket.TextMessage, []byte(m)); err != nil {
						log.Println("wsComms(): Error sending   message:", err.Error())
					}
				}
			}
		}(ws)

		send <- "Welcome..."

		// handle comms FROM the websocket
		for {
			_, bMsg, err := ws.ReadMessage() // In actual practice, we'd probably use ws.ReadJSON instead.
			if err != nil {
				log.Println("wsComms(): Error receiving message:", err.Error())
				break
			}
			send <- string(bMsg)
			log.Println("wsComms(): Received from client  : " + string(bMsg))
		}
	} else {
		log.Println("wsComms(): Error upgrading connection:", err)
	}
}
