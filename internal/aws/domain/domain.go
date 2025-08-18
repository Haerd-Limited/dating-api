package domain

import "mime/multipart"

type VoiceNoteUpload struct {
	AuthorID        string
	VoiceNoteHeader multipart.FileHeader
	VoiceNoteFile   multipart.File
	FolderPath      S3AudioFolderPath
}

type ImageUpload struct {
	AuthorID    string
	ImageHeader multipart.FileHeader
	ImageFile   multipart.File
	FolderPath  S3ImageFolderPath
}
