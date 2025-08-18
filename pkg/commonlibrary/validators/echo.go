package validators

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
)

func DecodeAndValidateEcho(r *http.Request) (*Request, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "multipart/form-data") {
		// Return empty request — all fields are optional
		return &Request{}, nil
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	caption := r.FormValue("caption")
	if caption == "" {
		return nil, ErrCaptionRequired
	}

	if len(caption) > 50 {
		return nil, ErrCaptionTooLong
	}

	captionPointer := &caption

	// Extract optional voice note
	var voiceNoteHeader *multipart.FileHeader

	var voiceNote multipart.File
	if file, header, err := r.FormFile("voice_note"); err == nil {
		voiceNote = file
		voiceNoteHeader = header
	} else if !errors.Is(err, http.ErrMissingFile) {
		return nil, fmt.Errorf("failed reading voice note: %w", err)
	}

	// Extract optional image
	var imageHeader *multipart.FileHeader

	var image multipart.File
	if file, header, err := r.FormFile("image"); err == nil {
		image = file
		imageHeader = header
	} else if !errors.Is(err, http.ErrMissingFile) {
		return nil, fmt.Errorf("failed reading image: %w", err)
	}

	var voiceNotePtr *multipart.File
	if voiceNote != nil {
		voiceNotePtr = &voiceNote
	}

	var imagePtr *multipart.File
	if image != nil {
		imagePtr = &image
	}

	return &Request{
		Caption:         captionPointer,
		VoiceNoteHeader: voiceNoteHeader,
		VoiceNoteFile:   voiceNotePtr,
		ImageHeader:     imageHeader,
		ImageFile:       imagePtr,
	}, nil
}
