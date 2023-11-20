package whisper_wrapper

import (
	"fmt"
	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
	"gwhisper_api/utils"
	"io"
	"strings"
)

// Whisper is the struct for the whisper model and all its methods
type Whisper struct {
	model whisper.Model
	// A channel to communicate with the model
	Com chan *utils.Com
}

// New creates a new Whisper struct
func New(modelpath string) *Whisper {
	model, err := whisper.New(modelpath)
	if err != nil {
		// panicking because we can't do anything without a model
		panic(err)
	}
	fmt.Println("Successfully loaded model: " + modelpath)
	// Create a new Whisper struct
	return &Whisper{model: model, Com: make(chan *utils.Com)}
}

// Predict will transcribe the wav file and return the result
func (w *Whisper) predict(com *utils.Com) (string, error) {
	// the result stringbuffer
	var buffer strings.Builder
	// Close the file handler when done
	defer com.Fh.Close()
	// Data variable to store the data from the wav file
	var data []float32
	// Callback to call from whisper
	var cb whisper.SegmentCallback
	// Create a new Model context
	context, err := w.model.NewContext()
	if err != nil {
		return "", err
	}

	// Set the context.language to de
	err = context.SetLanguage("de")
	if err != nil {
		return "", err
	}

	// Decode the WAV file - load the full buffer
	dec := wav.NewDecoder(com.Fh)
	if buf, err := dec.FullPCMBuffer(); err != nil {
		return "", err
	} else if dec.SampleRate != whisper.SampleRate {
		return "", fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	} else if dec.NumChans != 1 {
		return "", fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	} else {
		data = buf.AsFloat32Buffer().Data
	}
	// Process the data
	context.ResetTimings()
	if err := context.Process(data, cb, nil); err != nil {
		return "", err
	}

	// Get the strings out of the context
	for {
		segment, err := context.NextSegment()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		buffer.WriteString(segment.Text + " ")
	}
	return buffer.String(), nil
}

// Listen will listen for the next request from the channel
func (w *Whisper) Listen() {
	for c := range w.Com {
		result, err := w.predict(c)
		c.Result <- utils.Result{Result: result, Error: err}
	}
}
