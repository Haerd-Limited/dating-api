package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/interaction/dto"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	commonprofile "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard/dto"
)

func MapToSwipesResponse(result string) dto.SwipesResponse {
	return dto.SwipesResponse{
		Result: result,
	}
}

func MapToGetLikesResponse(domainLikes *domain.Likes) dto.GetLikesResponse {
	if domainLikes == nil {
		return dto.GetLikesResponse{
			Verified:   []dto.Like{},
			Unverified: []dto.Like{},
		}
	}

	var verified []dto.Like

	for _, domainLike := range domainLikes.Verified {
		like := dto.Like{
			Profile: commonprofile.ProfileCardToDto(domainLike.Profile),
			Message: &dto.Message{},
			Prompt:  &dto.Prompt{},
		}

		if domainLike.Prompt != nil {
			like.Prompt = &dto.Prompt{
				PromptID:      domainLike.Prompt.PromptID,
				Prompt:        domainLike.Prompt.Prompt,
				VoiceNoteURL:  domainLike.Prompt.VoiceNoteURL,
				CoverPhotoUrl: domainLike.Prompt.CoverPhotoUrl,
			}
		}

		if domainLike.Message != nil {
			like.Message.MessageText, like.Message.MessageType = domainLike.Message.MessageText, domainLike.Message.MessageType
		}

		verified = append(verified, like)
	}

	var unverified []dto.Like

	for _, domainLike := range domainLikes.Unverified {
		like := dto.Like{
			Profile: commonprofile.ProfileCardToDto(domainLike.Profile),
			Message: &dto.Message{},
			Prompt:  &dto.Prompt{},
		}

		if domainLike.Prompt != nil {
			like.Prompt = &dto.Prompt{
				PromptID:      domainLike.Prompt.PromptID,
				Prompt:        domainLike.Prompt.Prompt,
				VoiceNoteURL:  domainLike.Prompt.VoiceNoteURL,
				CoverPhotoUrl: domainLike.Prompt.CoverPhotoUrl,
			}
		}

		if domainLike.Message != nil {
			like.Message.MessageText, like.Message.MessageType = domainLike.Message.MessageText, domainLike.Message.MessageType
		}

		unverified = append(unverified, like)
	}

	return dto.GetLikesResponse{
		Verified:   verified,
		Unverified: unverified,
	}
}
