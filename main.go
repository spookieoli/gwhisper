package main

import (
	"gwhisper_api/webserver"
	"gwhisper_api/whisper_wrapper"
)

// System entry point
func main() {
	server := webserver.New("127.0.0.1", 8080)
	server.Whisper = whisper_wrapper.New("model/model_quant.bin")
	go server.Whisper.Listen()
	server.Start()
}

// to compile use: C_INCLUDE_PATH=/path/to/whisper.cpp LIBRARY_PATH=/path/to/whisper.cpp go build main.go
