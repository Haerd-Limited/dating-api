package mapper

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
)

const (
	minNameLen = 3
	maxNameLen = 20
)

var (
	ErrNameContainsSpaces = errors.New("name must not contain spaces")
	ErrInvalidNameLength  = errors.New("name must be between 3 and 20 characters")
	ErrInvalidID          = errors.New("id must be greater than 0")
)

func MapIntroRequestToDomain(request dto.IntroRequest, userID string) (domain.Intro, error) {
	firstName := strings.TrimSpace(request.FirstName)

	var lastName string
	if request.LastName != nil {
		lastName = strings.TrimSpace(*request.LastName)
	}

	if hasAnySpace(firstName) || hasAnySpace(lastName) {
		return domain.Intro{}, ErrNameContainsSpaces
	}

	// first name length check
	if l := len(firstName); l < minNameLen || l > maxNameLen {
		return domain.Intro{}, ErrInvalidNameLength
	}

	// last name length check
	if l := len(lastName); l < minNameLen || l > maxNameLen {
		return domain.Intro{}, ErrInvalidNameLength
	}

	if !looksLikeEmail(strings.TrimSpace(request.Email)) {
		return domain.Intro{}, commonErrors.ErrInvalidEmail
	}

	return domain.Intro{
		UserID:    userID,
		FirstName: firstName,
		LastName:  &lastName,
		// PhoneNumber: normalizePhone(request.PhoneNumber),
		Email: strings.TrimSpace(request.Email),
	}, nil
}

func MapBasicRequestToDomain(req dto.BasicsRequest, userID string) (domain.Basics, error) {
	dob, err := time.Parse("2006-01-02", req.Birthdate)
	if err != nil {
		return domain.Basics{}, commonErrors.ErrInvalidDob
	}

	if req.GenderID == 0 {
		return domain.Basics{}, ErrInvalidID
	}

	if req.DatingIntentionID == 0 {
		return domain.Basics{}, ErrInvalidID
	}

	return domain.Basics{
		UserID:            userID,
		Birthdate:         dob,
		HeightCm:          req.HeightCm,
		GenderID:          req.GenderID,
		DatingIntentionID: req.DatingIntentionID,
	}, nil
}

func MapLocationRequestToDomain(req dto.LocationRequest, userID string) (domain.Location, error) {
	return domain.Location{
		UserID:    userID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		City:      req.City,
		Country:   req.Country,
	}, nil
}

func MapLifestyleRequestToDomain(req dto.LifestyleRequest, userID string) (domain.Lifestyle, error) {
	if req.DrinkingID == 0 {
		return domain.Lifestyle{}, ErrInvalidID
	}

	if req.SmokingID == 0 {
		return domain.Lifestyle{}, ErrInvalidID
	}

	if req.MarijuanaID == 0 {
		return domain.Lifestyle{}, ErrInvalidID
	}

	if req.DrugsID == 0 {
		return domain.Lifestyle{}, ErrInvalidID
	}

	return domain.Lifestyle{
		UserID:      userID,
		DrinkingID:  req.DrinkingID,
		MarijuanaID: req.MarijuanaID,
		SmokingID:   req.SmokingID,
		DrugsID:     req.DrugsID,
	}, nil
}

func MapBackgroundRequestToDomain(req dto.BackgroundRequest, userID string) (domain.Background, error) {
	if req.EducationLevelID == 0 {
		return domain.Background{}, ErrInvalidID
	}

	if req.EthnicityID == 0 {
		return domain.Background{}, ErrInvalidID
	}

	return domain.Background{
		UserID:           userID,
		EducationLevelID: req.EducationLevelID,
		EthnicityID:      req.EthnicityID,
	}, nil
}

func MapBeliefsRequestToDomain(req dto.BeliefsRequest, userID string) (domain.Beliefs, error) {
	if req.PoliticalBeliefID == 0 {
		return domain.Beliefs{}, ErrInvalidID
	}

	if req.ReligionID == 0 {
		return domain.Beliefs{}, ErrInvalidID
	}

	return domain.Beliefs{
		UserID:             userID,
		PoliticalBeliefsID: req.PoliticalBeliefID,
		ReligionID:         req.ReligionID,
	}, nil
}

func MapLanguagesRequestToDomain(req dto.LanguagesRequest, userID string) (domain.Languages, error) {
	return domain.Languages{
		UserID:      userID,
		LanguageIDs: req.LanguageIDs,
	}, nil
}

func MapWorkAndEducationRequestToDomain(req dto.WorkAndEducationRequest, userID string) (domain.WorkAndEducation, error) {
	return domain.WorkAndEducation{
		UserID:     userID,
		Workplace:  req.Workplace,
		JobTitle:   req.JobTitle,
		University: req.University,
	}, nil
}

func MapPhotosRequestToDomain(req dto.PhotosRequest, userID string) (domain.UploadedPhotos, error) {
	var photos []domain.Photo

	for _, p := range req.UploadedPhotos {
		photos = append(photos, domain.Photo{
			URL:       p.URL,
			Position:  p.Position,
			IsPrimary: p.IsPrimary,
		})
	}

	return domain.UploadedPhotos{
		UserID: userID,
		Photos: photos,
	}, nil
}

func MapProfileToDomain(req dto.ProfileRequest, userID string) domain.Profile {
	return domain.Profile{
		UserID:               userID,
		ProfileBaseColour:    req.ProfileBaseColour,
		ProfileCoverPhotoURL: req.ProfileCoverPhotoURL,
	}
}

func MapPromptsRequestToDomain(req dto.PromptsRequest, userID string) (domain.Prompts, error) {
	var voicePrompts []domain.VoicePrompt

	for _, p := range req.UploadedPrompts {
		voicePrompts = append(voicePrompts, domain.VoicePrompt{
			URL:           p.URL,
			Position:      p.Position,
			IsPrimary:     p.IsPrimary,
			PromptType:    p.PromptType,
			CoverPhotoUrl: p.CoverPhotoUrl,
		})
	}

	return domain.Prompts{
		UserID:          userID,
		UploadedPrompts: voicePrompts,
	}, nil
}

// hasAnySpace returns true if s contains any Unicode whitespace character.
func hasAnySpace(s string) bool {
	return strings.IndexFunc(s, unicode.IsSpace) >= 0
}

func looksLikeEmail(s string) bool { return strings.Contains(s, "@") && strings.Contains(s, ".") }

/*func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	return s // stub
}
*/
