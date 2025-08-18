package domain

type (
	S3ImageFolderPath string
	S3AudioFolderPath string
)

const (
	FolderPostImages      S3ImageFolderPath = "post-images"
	FolderProfilePictures S3ImageFolderPath = "profile-pictures"
	FolderVoiceNotes      S3AudioFolderPath = "voice-notes"
)
