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

func MapToGetLikesResponse(domainLikes []domain.Like) dto.GetLikesResponse {
	var likes []dto.Like

	for _, domainLike := range domainLikes {
		like := dto.Like{
			Profile: commonprofile.ProfileCardToDto(domainLike.Profile),
			Message: &dto.Message{},
		}

		if domainLike.Message != nil {
			like.Message.MessageText, like.Message.MessageType = domainLike.Message.MessageText, domainLike.Message.MessageType
		}

		likes = append(likes, like)
	}

	return dto.GetLikesResponse{
		Likes: likes,
	}
}
