package webserver

import (
	"gwhisper_api/utils"
	"net/http"
	"os"
)

// Index is the route for the index page
func (ws *WebServer) Index(w http.ResponseWriter, r *http.Request) {
	ws.templates.ExecuteTemplate(w, "index.gohtml", nil)
}

// WavUpload is the route for uploading a wav file via post requestit will not use a template
func (ws *WebServer) WavUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// limit the max size of the request to 10 MB
		r.ParseMultipartForm(10 << 20)
		// Get the wavfile out of the request
		file, handler, err := r.FormFile("wavfile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// save the file to the filesystem - if there is not the same file already
		fn, err := utils.Utilizer.SaveFile(file, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Check if the File is a valid wav file
		if !utils.Utilizer.IsWavFile("upload/" + fn) {
			// Delete the file if it is not a valid wav file
			utils.Utilizer.DeleteFile("upload/" + fn)
			// Send the user that the file is not a valid wav file
			http.Error(w, "The file you uploaded is not a valid wav file", http.StatusInternalServerError)
		}
		// Convert the file to mono / 16 bit if necessary
		err = utils.Utilizer.ConvertWavFile(fn)
		if err != nil {
			// Delete the file if it is not a valid wav file
			utils.Utilizer.DeleteFile("upload/" + fn)
			// Send the user that the file is not a valid wav file
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// get a transcript from the wav file
		fh, err := os.Open("upload/converted/" + fn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer fh.Close()
		// Create the transcript
		com := &utils.Com{Fh: fh, Result: make(chan utils.Result)}
		ws.Whisper.Com <- com
		result := <-com.Result
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		// Send the transcript to the user as json
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"transcript\": \"" + result.Result + "\"}"))
		return
	}
	// Respond with a 404 if the method is not POST
	http.Error(w, "404 page not found", http.StatusNotFound)
}
