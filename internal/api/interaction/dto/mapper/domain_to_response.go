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
			FreeToMatch: []dto.Like{},
			SlotsFull:   []dto.Like{},
		}
	}

	return dto.GetLikesResponse{
		FreeToMatch:        mapLikes(domainLikes.FreeToMatch),
		SlotsFull:          mapLikes(domainLikes.SlotsFull),
		ActiveMatchesCount: domainLikes.ActiveMatchesCount,
		MatchSlotLimit:     domainLikes.MatchSlotLimit,
	}
}

func mapLikes(domainLikes []domain.Like) []dto.Like {
	if len(domainLikes) == 0 {
		return []dto.Like{}
	}

	likes := make([]dto.Like, 0, len(domainLikes))

	for _, domainLike := range domainLikes {
		like := dto.Like{
			Profile:            commonprofile.ProfileCardToDto(domainLike.Profile),
			Message:            &dto.Message{},
			Prompt:             &dto.Prompt{},
			TargetAtMatchLimit: domainLike.TargetAtMatchLimit,
			IsFavourited:       domainLike.IsFavourited,
		}

		if domainLike.Prompt != nil {
			like.Prompt = &dto.Prompt{
				PromptID:              domainLike.Prompt.PromptID,
				Prompt:                domainLike.Prompt.Prompt,
				VoiceNoteURL:          domainLike.Prompt.VoiceNoteURL,
				CoverMediaURL:         domainLike.Prompt.CoverMediaURL,
				CoverMediaType:        domainLike.Prompt.CoverMediaType,
				CoverMediaAspectRatio: domainLike.Prompt.CoverMediaAspectRatio,
			}
		}

		if domainLike.Message != nil {
			like.Message.MessageText = domainLike.Message.MessageText
			like.Message.MessageType = domainLike.Message.MessageType
			like.Message.VoiceNoteURL = domainLike.Message.VoiceNoteURL
			like.Message.MediaSeconds = domainLike.Message.MediaSeconds
		}

		likes = append(likes, like)
	}

	return likes
}
