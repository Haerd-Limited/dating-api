package domain

//todo:refactor
type GetFollowingResult struct {
	Following []*User
}

type GetFollowersResult struct {
	Followers []*User
}
