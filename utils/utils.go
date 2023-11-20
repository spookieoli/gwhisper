package utils

import (
	"crypto/sha1"
	"fmt"
	"github.com/go-audio/audio"
	wav "github.com/go-audio/wav"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

// Result is a struct to store the result of the model
type Result struct {
	// The result string
	Result string
	// The error if there is one
	Error error
}

// Com is the struct to communicate with the model
type Com struct {
	// a file handler to a wave file
	Fh *os.File
	// a channel to get the result from the model
	Result chan Result
}

// Utils is the struct for the utils package
type Utils struct {
}

var Utilizer *Utils

// init initializes the utils package
func init() {
	Utilizer = &Utils{}
}

// Savefile will save a file to the filesystem with the filename written as sha1 hash
func (u *Utils) SaveFile(file multipart.File, filename string) (string, error) {
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	//  the filename is the hash
	fn := fmt.Sprintf("%x", hash.Sum(nil))

	// Check if the file already exists
	if !u.CheckFileExists("upload/" + fn + strings.Split(filename, ".")[1]) {
		// seek back to the beginning of the file
		if _, err := file.Seek(0, 0); err != nil {
			return "", err
		}
		// create the file with the hash as filename
		dst, err := os.Create("upload/" + fn + "." + strings.Split(filename, ".")[1])
		if err != nil {
			return "", err
		}
		defer dst.Close()
		// Save the file to the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)) + "." + strings.Split(filename, ".")[1], nil
}

// IsWavFile checks if a file is a wav file
func (u *Utils) IsWavFile(filename string) bool {
	fh, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer fh.Close()
	// Try to decode the file
	decoder := wav.NewDecoder(fh)
	if decoder == nil {
		return false
	}
	return true
}

// DeleteFile deletes a file from the filesystem
func (u *Utils) DeleteFile(filename string) error {
	return os.Remove(filename)
}

// ConvertWavFile will - if necessary - change a wav file in terms of being mono and 16 bit
func (u *Utils) ConvertWavFile(filename string) error {
	// Open the file
	fh, err := os.Open("upload/" + filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Set up WAV decoder
	decoder := wav.NewDecoder(fh)
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return err
	}

	// Check if the file is already mono
	var monoBuf *audio.IntBuffer
	if buf.Format.NumChannels > 1 {
		// Convert to mono
		monoBuf = &audio.IntBuffer{Data: make([]int, len(buf.Data)/buf.Format.NumChannels), Format: &audio.Format{SampleRate: buf.Format.SampleRate, NumChannels: 1}}
		for i := 0; i < len(monoBuf.Data); i++ {
			sum := 0
			for ch := 0; ch < buf.Format.NumChannels; ch++ {
				sum += buf.Data[i*buf.Format.NumChannels+ch]
			}
			monoBuf.Data[i] = sum / buf.Format.NumChannels
		}
	} else {
		// Use the file as is if it's already mono
		monoBuf = buf
	}

	// Save the mono WAV file in 16-bit
	outFile, err := os.Create("upload/converted/" + filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	encoder := wav.NewEncoder(outFile, monoBuf.Format.SampleRate, 16, monoBuf.Format.NumChannels, 1)
	if err := encoder.Write(monoBuf); err != nil {
		return err
	}
	return encoder.Close()
}

// CheckFileExists checks if a file exists
func (u *Utils) CheckFileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
