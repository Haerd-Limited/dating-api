package domain

import (
	"mime/multipart"
	"time"
)

type (
	User struct {
		ID             string
		Email          string
		HashedPassword string
		Dob            time.Time
	}

	UserProfile struct {
		ID              string
		Username        string
		FullName        string
		Bio             *string
		Gender          *string
		DateOfBirth     *time.Time
		FollowerCount   int
		FollowingCount  int
		PostCount       int
		IsFollowing     bool
		CreatedAt       time.Time
		ProfileImageURL *string
	}

	FollowUser struct {
		FollowingID string
		FollowerID  string
	}

	UnfollowUser = FollowUser

	ViewProfile struct {
		TargetUserID string
		ViewerID     string
	}

	UpdateUserProfile struct {
		UserID       string
		Username     string
		FullName     string
		DateOfBirth  *time.Time
		Biography    string
		Gender       string
		Email        string
		ProfileImage *multipart.File
		ImageHeader  *multipart.FileHeader
	}
)
