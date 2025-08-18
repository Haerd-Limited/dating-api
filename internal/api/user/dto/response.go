package dto

import "time"

type GetFollowingResponse struct {
	Message   string         `json:"message"`
	Following []*UserDetails `json:"following,omitempty"`
}

type UpdateUserProfileResponse struct {
	Message     string       `json:"message"`
	UserDetails *UserDetails `json:"user_details,omitempty"`
}

type GetFollowersResponse struct {
	Message   string         `json:"message"`
	Followers []*UserDetails `json:"followers,omitempty"`
}

type GetProfileResponse struct {
	Message string       `json:"message"`
	Profile *UserProfile `json:"profile,omitempty"`
}

type UserProfile struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	FullName       string     `json:"full_name"`
	Bio            *string    `json:"bio,omitempty"`
	Gender         *string    `json:"gender,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	FollowerCount  int        `json:"follower_count"`
	FollowingCount int        `json:"following_count"`
	IsFollowing    bool       `json:"is_following"`
	PostCount      int        `json:"post_count"`
	CreatedAt      time.Time  `json:"created_at"`
	ProfilePicURL  *string    `json:"profile_pic_url,omitempty"`
}

type UserDetails struct {
	ID            *string    `json:"id,omitempty"`
	Username      *string    `json:"username,omitempty"`
	Email         *string    `json:"email,omitempty"`
	FullName      *string    `json:"full_name,omitempty"`
	Bio           *string    `json:"bio,omitempty"`
	Gender        *string    `json:"gender,omitempty"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	ProfilePicURL *string    `json:"profile_pic_url,omitempty"`
}
