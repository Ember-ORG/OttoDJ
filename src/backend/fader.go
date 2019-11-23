package main

// Holds fadeItter and trackItter
import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	id3 "github.com/mikkyang/id3-go"
)

/*
fadeItter - Is used to fade in and out
trackItter - Represents the position into a song
*/
var floatPersistence map[int][]float64

var stringPersistence map[int][]string

var stream beep.StreamSeeker

var ctrl *beep.Ctrl

var volume *effects.Volume

var sampleRate float64

var seconds float64

var format beep.Format

var skipAmount int

var milliRate = 1

var totalSongs int

var audioSoFar float64

// fader is a type so that fader.Stream() can be used with proper parameters to run properly
type fader struct {
	// Streamer to fade
	Streamer beep.Streamer
	// How long in samples to fade in, and to fade out
	TimeSpan float64
	// What the volume should be for the streamer
	Volume float64
	// SampleRate of Streamer
	SampleRate float64
	// How long the audio is, so that fading in and out works properly
	audioLength float64
	// ID so that it can persist itterators between bits of slices
	id int
	// Metadata
	metadata id3.Tagger
	// Determines whether or not it should send metadata and other ways it should perform
	buffer bool
}

// For testing fading capabilities
func init() {
	// Necessary for floatPersistence map, otherwise there is a nil map error
	floatPersistence = make(map[int][]float64)
	// Necessary for stringPersistence map, otherwise there is a nil map error
	stringPersistence = make(map[int][]string)
}

/*
var floatPersistence map[int][]float64

var ctrl *beep.Ctrl

var volume *effects.Volume

var sampleRate float64

var seconds float64

var format beep.Format

var skipAmount int

var milliRate = 1

var firstPlay = true
*/
// Crossfades between all songs specified in files
func streamCreater(files ...string) {
	floatPersistence = make(map[int][]float64)
	volume = nil
	sampleRate = 0
	seconds = 0
	format = beep.Format{}
	skipAmount = 0
	milliRate = 1
	// Streamer that will contain all files
	var streamer beep.Streamer
	// Create 1000 samples of silence so that beep.Mix has a non-nil streamer to work with
	streamer = beep.Silence(1000)
	// The time span of the file previous to the one calculating on it. Used to get timing for crossfading right
	var lastTimeSpan float64
	// Specifies how long the streamer is, so that timing for crossfading is correct
	var position float64
	// Used so that speaker.Init has valid values for SampleRate, etc. Probably not a good idea if the SampleRates are different between files
	// Iterate through all files specified to add them to streamer with proper crossfade
	var quit bool
	for id, name := range files {
		// Open the file
		f, err := os.Open(name)
		defer f.Close()
		if err != nil {
			fmt.Println("Couldn't find file " + name)
		}
		// Declared here so that format isn't specific to this block
		var s beep.StreamSeekCloser
		// Decode the file
		s, format, err = mp3.Decode(f)
		if err != nil {
			fmt.Println("Please ensure that " + name + " is an mp3, or is not corrupted")
			quit = true
			return
		}

		mp3File, err := id3.Open(name)
		if mp3File.Title() == "" {
			mp3File.SetTitle(strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)))
		}
		var faderStream = &fader{Streamer: s, Volume: 1, SampleRate: float64(format.SampleRate), TimeSpan: float64(format.SampleRate.N(time.Second * 9)), audioLength: float64(s.Len()), id: id, metadata: mp3File.Tagger, buffer: true}
		// Create streamer with fading applied
		changedStreamer := beep.StreamerFunc(faderStream.Stream)
		// Create amount of silence before playing sound. Uses position, which by itself would make it play after the previous song. Subtracting lastTimeSpan makes a crossfade effect
		silenceAmount := int(position - lastTimeSpan)
		// Keeps previous streamer, and adds the new streamer with the silence in the beginning so it doesn't play over other songs
		streamer = beep.Mix(streamer, beep.Seq(beep.Silence(silenceAmount), changedStreamer))
		// Add position for next file
		position = position + faderStream.audioLength
		// Set last time span to current time span for next file
		lastTimeSpan = faderStream.TimeSpan
		totalSongs = id + 1
	}
	if quit {
		return
	}
	err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*2))
	if err != nil {
		fmt.Println("Couldn't initialize the speaker" + err.Error())
	}
	buff := beep.NewBuffer(format)
	buff.Append(streamer)
	stream = buff.Streamer(0, buff.Len())
	ctrl = &beep.Ctrl{Streamer: stream}
	go w.SendMessage("unlod", func(m *astilectron.EventMessage) {
	})
	sampleRate = float64(format.SampleRate)
	// Initialize speaker
	done := speakerPlay(ctrl)
	speaker.UnderrunCallback(func() {
		fmt.Println("Underrun detected!")
	})
	<-done
}

func speakerPlay(streamer beep.Streamer) chan struct{} {
	fmt.Println("Playing stream!")
	// Create done channel so that program doesn't exit before all songs are played
	done := make(chan struct{})
	// Play streamer (doesn't belong here)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	go func() {
		var first = true
		var songPosition = 0
		var i = 0
		for true {
			songPosition++
			go w.SendMessage("s"+strconv.Itoa(songPosition), func(m *astilectron.EventMessage) {
			})
			fmt.Println(i, stream.Position(), floatPersistence[i][2])
			if floatPersistence[i][1] >= floatPersistence[i][2] || first {
				go w.SendMessage("m"+strconv.Itoa(stream.Len()/int(format.SampleRate)), func(m *astilectron.EventMessage) {
				})
				if len(stringPersistence[i]) == 2 {
					if stringPersistence[i][0] != "" && stringPersistence[i][1] != "" {
						go w.SendMessage("1a"+stringPersistence[i][0], func(m *astilectron.EventMessage) {
						})
						go w.SendMessage("1t"+stringPersistence[i][1], func(m *astilectron.EventMessage) {
						})
					} else {
						fmt.Println("Metadata is nil")
					}
				} else {
					fmt.Println("Metadata is nil")
				}
				songPosition = 0
				if !first {
					i++
				}
				first = false

			}
			time.Sleep(time.Second)
		}
	}()
	return done
}

/*
	ctrl = &beep.Ctrl{Streamer: volume}
	sampleRate = float64(format.SampleRate)
	// Initialize speaker
	done := speakerPlay(format.SampleRate, format.SampleRate.N(time.Millisecond*time.Duration(milliRate)))
	speaker.UnderrunCallback(func() {
		fmt.Println("Underrun")
		milliRate = milliRate + 50
		speakerPlay(format.SampleRate, format.SampleRate.N(time.Millisecond*time.Duration(milliRate)))
	})
	<-done
}

func speakerPlay(SampleRate beep.SampleRate, BufferSize int) chan struct{} {
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*time.Duration(milliRate)))
	// Create done channel so that program doesn't exit before all songs are played
	done := make(chan struct{})
	// Play streamer (doesn't belong here)
	fmt.Println("Ill")
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		fmt.Println("Ill")
		close(done)
	})))
	return done
}
*/

// Stream edits streamer so that it fades
func (v *fader) Stream(samples [][2]float64) (n int, ok bool) {
	// Determines if this specific streamer has been run before. If it hasn't then it needs to create fadeItter and trackItter for it
	if len(floatPersistence) < v.id+1 {
		// Print ID of song
		fmt.Println(v.id)
		// Create fadeItter and trackItter for the ID, and assign them to defaults of 0
		floatPersistence[v.id] = []float64{0, 0, v.audioLength}
		stringPersistence[v.id] = []string{"", ""}
		go func() {
			// If it is a buffer, then send a message to begin loading screen
			if v.buffer {
				go w.SendMessage("lod", func(m *astilectron.EventMessage) {
				})
			}
			if v.metadata != nil {
				stringPersistence[v.id][0] = v.metadata.Artist()
				stringPersistence[v.id][1] = v.metadata.Title()
			}
		}()
	}
	// Assign name to the map's ints for easier reading
	/*
		fadeItter - Is used to fade in and out
		trackItter - Represents the position into a song
	*/
	var fadeItter = &floatPersistence[v.id][0]
	var trackItter = &floatPersistence[v.id][1]
	// Use default streamer, and revise off of that
	n, ok = v.Streamer.Stream(samples)
	if !ok {
		fmt.Println("Streaming failure")
	}
	var gain float64
	gain = v.Volume
	// x1 is 0 and represents the start of the fade
	var x1 float64
	// The start of the fade should be silent, so y1 is 0
	var y1 float64
	// End point should be the TimeSpan set so that at the end of the TimeSpan, the gain is at requested value
	var x2 = v.TimeSpan
	// The requested gain, which will be played at the end of the TimeSpan
	var y2 = gain
	// Create the slope for a line representing this
	slopeUp := slopeCalc(x1, y1, x2, y2)
	//slopeDown := slopeCalc(x1, y2, x2, y1)
	// By default, sampleGain is the requested gain so between fadepoints, it is normal
	var sampleGain = gain
	// For each recieved sample, apply fade to it if necessary
	for i := range samples[:n] {
		if math.Mod(*trackItter, v.SampleRate) == 0 {
			seconds = *trackItter / v.SampleRate
			// If it is a buffer, then just send load progress
			if v.buffer {
				go w.SendMessage("prog "+strconv.Itoa(v.id+1)+" "+strconv.Itoa(int((seconds/(v.audioLength/v.SampleRate)*100)))+" "+strconv.Itoa(totalSongs), func(m *astilectron.EventMessage) {
				})
			} else if len(floatPersistence) < v.id+2 {
				go w.SendMessage("s"+strconv.Itoa(int(seconds)), func(m *astilectron.EventMessage) {
				})
			}
		}
		// If the position in the track is after or at the time where it should begin to fade, then fade
		if *trackItter >= v.audioLength-v.TimeSpan {
			// Slope-intercept form to get gain
			/*
				m					x 							+ 	b
				Calculated slope	The position in the fade		The y intercept of the gain, so that it fades down from the gain
			*/
			sampleGain = -(slopeUp * float64(*fadeItter)) + gain
			// Increment fade so that the next iteration will reduce the gain by more
			*fadeItter++
			// Prevents possible bug where the gain may become negative, which will result in the song's gain becoming high again
			if sampleGain < 0 {
				sampleGain = 0
			}
			// If the position of the track is before the specified TimeSpan, and the fadeItter isn't above the TimeSpan, begin to fade in.
		} else if *trackItter <= v.TimeSpan && slopeUp*float64(*fadeItter) <= gain {
			// Slope-intercept form to get gain
			/*
				m					x 							+ 	b
				Calculated slope	The position in the fade		0, because it is fading in from nothing
			*/
			sampleGain = slopeUp * float64(*fadeItter)
			// Increment fade so that the next iteration will reduce the gain by more
			*fadeItter++
		} else {
			// Ensures fadeItter isn't already high from fading in when it is time to fade out
			*fadeItter = 0
		}
		// Set the samples to the calculated gain
		samples[i][0] *= sampleGain
		samples[i][1] *= sampleGain
		// Increment trackItter to update position in track

		*trackItter++
	}
	// Return the samples with gain applied, and whether or not operations were successful
	return n, ok
}

// Calculates the slope between two points
func slopeCalc(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	return (y2 - y1) / (x2 - x1)
}

func pop(num int, s beep.Streamer) beep.Streamer {
	return &popStruct{
		s:             s,
		startingPoint: num,
	}
}

type popStruct struct {
	s             beep.Streamer
	startingPoint int
}

func (t *popStruct) Stream(samples [][2]float64) (n int, ok bool) {
	startingPoint := t.startingPoint
	fmt.Println(startingPoint, len(samples))
	n, ok = t.s.Stream(samples[startingPoint:])
	t.startingPoint += len(samples)
	return n, ok
}

func (t *popStruct) Err() error {
	return t.s.Err()
}
