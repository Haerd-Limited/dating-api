package validators

import (
	"errors"
	"fmt"
	"net/http"
)

func DecodeAndValidateComment(r *http.Request) (*Request, error) {
	// Parse multipart form with a maxMemory size (e.g., 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	// Extract title
	var captionPointer *string

	caption := r.FormValue("caption")
	if caption == "" {
		captionPointer = nil
		// return nil, ErrCaptionRequired
	} else {
		captionPointer = &caption

		if len(caption) > 50 {
			return nil, ErrCaptionTooLong
		}
	}

	// Extract voice note file
	voiceNote, voiceNoteHeader, err := r.FormFile("voice_note")
	if err != nil {
		return nil, ErrVoiceNoteRequired
	}

	defer func() {
		if err := voiceNote.Close(); err != nil {
			// Log the error but don’t fail the request — it's deferred.
			fmt.Println("failed to close voiceNote:", err)
		}
	}()

	// (Optional) Extract image file
	image, imageHeader, err := r.FormFile("image") // don't error out if missing
	if errors.Is(err, http.ErrMissingFile) {
		image = nil
		imageHeader = nil
	}

	return &Request{
		Caption:         captionPointer,
		VoiceNoteHeader: voiceNoteHeader,
		VoiceNoteFile:   &voiceNote,
		ImageHeader:     imageHeader,
		ImageFile:       &image,
	}, nil
}
