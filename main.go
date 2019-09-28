package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/pkg/errors"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: foo FILE.wav...")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	// Init speaker.
	initStart := time.Now()
	const sampleRate beep.SampleRate = 22050
	if err := initSpeaker(sampleRate); err != nil {
		log.Fatalf("%+v", err)
	}
	log.Printf("initializing speakers took %v", time.Since(initStart))
	for _, wavPath := range flag.Args() {
		// Decode WAV file.
		decodeStart := time.Now()
		audioBuf, err := decode(wavPath)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		log.Printf("decoding %q took %v", wavPath, time.Since(decodeStart))
		fmt.Println()
		for i := 0; i < 10; i++ {
			// Add playback control to audio buffer.
			ctrlStart := time.Now()
			ctrl := &beep.Ctrl{
				Streamer: audioBuf.Streamer(0, audioBuf.Len()),
			}
			log.Printf("creating ctrl took %v", time.Since(ctrlStart))
			// Play sound.
			playStart := time.Now()
			done := play(ctrl)
			<-done
			log.Printf("playing sound of length %v took %v", audioBuf.Len(), time.Since(playStart))
			fmt.Println()
		}
	}
}

func initSpeaker(sampleRate beep.SampleRate) error {
	bufferSize := sampleRate.N(time.Second / 10)
	if err := speaker.Init(sampleRate, bufferSize); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func play(ctrl *beep.Ctrl) <-chan struct{} {
	playStart := time.Now()
	first := beep.Callback(func() {
		log.Printf("time taken until first sound: %v", time.Since(playStart))
	})
	done := make(chan struct{})
	last := beep.Callback(func() {
		done <- struct{}{}
	})
	streamer := beep.Seq(first, ctrl, last)
	speaker.Play(streamer)
	return done
}

func decode(wavPath string) (*beep.Buffer, error) {
	buf, err := ioutil.ReadFile(wavPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r := bytes.NewReader(buf)
	streamer, format, err := wav.Decode(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	audioBuf := beep.NewBuffer(format)
	audioBuf.Append(streamer)
	if err := streamer.Close(); err != nil {
		return nil, errors.WithStack(err)
	}
	return audioBuf, nil
}
