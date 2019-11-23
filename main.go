package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	astilectron "github.com/asticode/go-astilectron"
	"github.com/gin-gonic/gin"
)

type bind struct {
	Value int `json:"value"`
}

var w *astilectron.Window

func main() {
	/*
		w := webview.New(webview.Settings{
			Title:     "OttoDJ",
			URL:       "http://127.0.0.1:8080",
			Debug:     true,
			Resizable: true,
		})
		defer w.Exit()
	*/
	go func() {
		router := gin.Default()
		router.StaticFS("/", assetFS())
		router.Run(":8080")
	}()
	/*
		w.Bind("counter", &bind{})
		w.Run()
	*/
	// Initialize astilectron
	a, err := astilectron.New(astilectron.Options{
		AppName:            "OttoDJ",
		AppIconDefaultPath: "resources/app/images/favicon.png",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	// Start astilectron
	a.Start()

	// Create a new window
	w, err = a.NewWindow("http://127.0.0.1:8080/app", &astilectron.WindowOptions{
		Center: astilectron.PtrBool(true),
		Height: astilectron.PtrInt(384),
		Width:  astilectron.PtrInt(550),
		Frame:  astilectron.PtrBool(false),
		//Resizable:      astilectron.PtrBool(false),
		HasShadow:      astilectron.PtrBool(true),
		Transparent:    astilectron.PtrBool(true),
		WebPreferences: &astilectron.WebPreferences{AllowRunningInsecureContent: astilectron.PtrBool(true), WebSecurity: astilectron.PtrBool(true)},
	})
	if err != nil {
		log.Fatal(err)
	}
	w.Create()
	w.OpenDevTools()
	// This will listen to messages sent by Javascript
	w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		// Unmarshal
		var s string
		m.Unmarshal(&s)

		if strings.HasPrefix(s, "p") {
			s = strings.TrimLeft(s, "p")
			sliceMsg := strings.Split(s, ",")
			streamCreater(sliceMsg...)
		} else if strings.HasPrefix(s, "stp") {
			w.Close()
			a.Close()
			a.Quit()
			a.Stop()
		} else if strings.HasPrefix(s, "0") {
			if ctrl != nil {
				if ctrl.Paused {
					ctrl.Paused = false
				} else {
					ctrl.Paused = true
				}
			} else {
				fmt.Println("Tried to play, but no music loaded")
			}
		} else if strings.HasPrefix(s, "v") {
			s = strings.TrimLeft(s, "v")
			vol, err := strconv.Atoi(s)
			if err != nil {
				fmt.Println("Recieved non-int volume")
			}
			if volume != nil {
				if vol == 0 {
					volume.Silent = true
				} else {
					volume.Silent = false
					volume.Base = float64((vol / 100))
				}
			} else {
				fmt.Println("Tried to set volume, but no music loaded")
			}
		} /*else if strings.HasPrefix(s, "l") {
			s = strings.TrimLeft(s, "l")
			loc, err := strconv.Atoi(s)
			if err != nil {
				fmt.Println("Location wasn't an int!")
			}
			fmt.Println("Must skip to", loc, sampleRate)
			b := beep.NewBuffer(format)
			b.Append(ctrl)
			var newStream beep.StreamSeeker
			if seconds > sampleRate {
				newStream = b.Streamer(0, loint(seconds*sampleRate))
			} else {
				newStream = b.Streamer(int(float64(loc)*sampleRate), b.Len())
			}
			speaker.Lock()
			ctrl.Streamer = newStream
			speaker.Unlock()
		}*/
		return nil
	})
	// Blocking pattern
	a.Wait()
}
