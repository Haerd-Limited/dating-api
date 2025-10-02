package messages

// Author-friendly messages sent to the FE
const (
	InternalServerErrorMsg      = "Something went wrong that is not your fault. Please try again later."
	VoiceNoteTooLongMsg         = "Voice note too long. Please upload a recording under 1 minute."
	VoiceNoteTooShortMsg        = "Voice note too short. Please record at least 5 seconds."
	VoiceNoteRequiredMsg        = "Please provide a voicenote"
	CaptionRequiredMsg          = "Please provide a caption"
	CaptionTooLongMsg           = "Caption too long. It must be 50 characters or less"
	NotAnEchoMsg                = "This is not an echo"
	UnAuthorisedActionMsg       = "You’re not allowed to perform this action"
	InvalidDobMsg               = "Please enter your date of birth in the format YYYY-MM-DD (e.g. 1995-08-24)."
	InvalidGenderMsg            = "Please select male or female."
	InvalidUploadFormMsg        = "Upload failed. Please use a valid form to upload files."
	EmailAlreadyExistsMsg       = "Email already exists."
	UserDetailsAlreadyExistsMsg = "An account with these details already exists."
	UserNameAlreadyExistsMsg    = "Username already exists."
	InvalidUUIDMsg              = "Invalid ID format. Must be a valid UUID."
	InvalidIDMsg                = "Invalid ID provided"
	AuthenticationRequiredMsg   = "Authentication required: please log in"
	RequestTimeoutMsg           = "Request timed out"
	RequestCancelledMsg         = "Request cancelled"
	AllFieldsRequiredMsg        = "All fields are required"
	SocialsNotAllowedMsg        = "To keep Haerd personality-first, socials aren’t allowed in your display name, job title, work place or university. Please remove @handles, links, or “dot com” text."
)
