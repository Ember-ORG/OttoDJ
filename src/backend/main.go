package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	astilectron "github.com/asticode/go-astilectron"
	"github.com/faiface/beep/speaker"
	"github.com/gin-gonic/gin"
)

var w *astilectron.Window

// Change this for production! This activates developer features throughout the program
var devel = true

func main() {
	resizeable := false
	if devel {
		fmt.Println("Developer mode is enabled!")
		fmt.Println("Please be aware that this could cause bugs or other undesirable side effects that we are not responsible for")
		fmt.Println("You can access a website version of the website at 127.0.0.1:8080")
		resizeable = true
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	go func() {
		router := gin.Default()
		router.StaticFS("/", assetFS())
		err := router.Run(":8080")
		if err != nil {
			log.Fatal(err)
		}
	}()

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
	err = a.Start()
	if err != nil {
		log.Fatal(err)
	}
	height := 384
	width := 550
	if devel {
		height = 1080
		width = 1920
	}

	// Create a new window
	w, err = a.NewWindow("http://127.0.0.1:8080/app", &astilectron.WindowOptions{
		Center:    astilectron.PtrBool(true),
		Height:    astilectron.PtrInt(height),
		Width:     astilectron.PtrInt(width),
		Frame:     astilectron.PtrBool(false),
		Resizable: astilectron.PtrBool(resizeable),
	})
	if err != nil {
		log.Fatal(err)
	}
	err = w.Create()
	if err != nil {
		log.Fatal(err)
	}
	if devel {
		err = w.OpenDevTools()
		if err != nil {
			log.Fatal(err)
		}
	}
	// This will listen to messages sent by Javascript
	w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		// Unmarshal
		var s string
		err = m.Unmarshal(&s)
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasPrefix(s, "p") {
			s = strings.TrimLeft(s, "p")
			sliceMsg := strings.Split(s, ",")
			streamCreater(sliceMsg...)
		} else if strings.HasPrefix(s, "stp") {
			err = w.Close()
			if err != nil {
				log.Fatal(err)
			}
			a.Close()
			err = a.Quit()
			if err != nil {
				log.Fatal(err)
			}
			a.Stop()
		} else if strings.HasPrefix(s, "0") {
			if ctrl != nil {
				ctrl.Paused = true
			} else {
				fmt.Println("Thou hast not loaded thy music")
			}
			fmt.Println("Stopping...")
		} else if strings.HasPrefix(s, "1") {
			if ctrl != nil {
				ctrl.Paused = false
			} else {
				fmt.Println("Thou hast not loaded thy music")
			}
			fmt.Println("Playing...")
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
		} else if strings.HasPrefix(s, "l") {
			s = strings.TrimLeft(s, "l")
			loc, err := strconv.Atoi(s)
			if err != nil {
				fmt.Println("Location wasn't an int!")
			}
			fmt.Println("Must skip to", loc, sampleRate, int(float64(loc)*(sampleRate)))
			newStreamer := pop(int(float64(loc)*(sampleRate)), ctrl.Streamer)
			speaker.Lock()
			ctrl.Streamer = newStreamer
			speaker.Unlock()
		}
		return nil
	})
	// Blocking pattern
	a.Wait()
}
