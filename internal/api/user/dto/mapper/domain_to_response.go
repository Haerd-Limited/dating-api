package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/user/dto"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
)

func ToUpdateProfileResponse(result *domain.User, message string) *dto.UpdateUserProfileResponse {
	if result == nil {
		return &dto.UpdateUserProfileResponse{
			Message: message,
		}
	}

	return &dto.UpdateUserProfileResponse{
		Message:     message,
		UserDetails: ToUserDetails(result),
	}
}

func ToGetFollowingResponse(result *domain.GetFollowingResult, message string) *dto.GetFollowingResponse {
	if result == nil {
		return &dto.GetFollowingResponse{
			Message: message,
		}
	}

	var following []*dto.UserDetails

	for _, u := range result.Following {
		if u == nil {
			continue
		}

		user := ToUserDetails(u)
		following = append(following, user)
	}

	return &dto.GetFollowingResponse{
		Message:   message,
		Following: following,
	}
}

func ToGetFollowersResponse(result *domain.GetFollowersResult, message string) *dto.GetFollowersResponse {
	if result == nil {
		return &dto.GetFollowersResponse{
			Message: message,
		}
	}

	var following []*dto.UserDetails

	for _, u := range result.Followers {
		if u == nil {
			continue
		}

		user := ToUserDetails(u)
		following = append(following, user)
	}

	return &dto.GetFollowersResponse{
		Message:   message,
		Followers: following,
	}
}

func ToGetProfileResponse(result *domain.UserProfile, message string) *dto.GetProfileResponse {
	if result == nil {
		return &dto.GetProfileResponse{
			Message: message,
		}
	}

	return &dto.GetProfileResponse{
		Message: message,
		Profile: ToUserProfile(result),
	}
}

func ToGetMyProfileResponse(result *domain.UserProfile, message string) *dto.GetProfileResponse {
	if result == nil {
		return &dto.GetProfileResponse{
			Message: message,
		}
	}

	return &dto.GetProfileResponse{
		Message: message,
		Profile: ToMyProfile(result),
	}
}

func ToUserDetails(user *domain.User) *dto.UserDetails {
	if user == nil {
		return nil
	}

	return &dto.UserDetails{
		ID:            &user.ID,
		Username:      &user.Username,
		Email:         &user.Email,
		FullName:      &user.FullName,
		Bio:           &user.Bio,
		Gender:        &user.Gender,
		DateOfBirth:   &user.Dob,
		ProfilePicURL: user.ProfileImageURL,
	}
}

func ToUserProfile(user *domain.UserProfile) *dto.UserProfile {
	if user == nil {
		return nil
	}

	return &dto.UserProfile{
		ID:             user.ID,
		Username:       user.Username,
		FullName:       user.FullName,
		Bio:            user.Bio,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		IsFollowing:    user.IsFollowing,
		CreatedAt:      user.CreatedAt,
		ProfilePicURL:  user.ProfileImageURL,
		PostCount:      user.PostCount,
	}
}

func ToMyProfile(user *domain.UserProfile) *dto.UserProfile {
	if user == nil {
		return nil
	}

	return &dto.UserProfile{
		ID:             user.ID,
		Username:       user.Username,
		FullName:       user.FullName,
		Bio:            user.Bio,
		Gender:         user.Gender,
		DateOfBirth:    user.DateOfBirth,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		IsFollowing:    user.IsFollowing,
		CreatedAt:      user.CreatedAt,
		ProfilePicURL:  user.ProfileImageURL,
		PostCount:      user.PostCount,
	}
}
