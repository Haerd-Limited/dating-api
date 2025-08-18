package algorithm

func CalculateScore(likeCount, commentCount, echoCount int) *int {
	score := (likeCount * 2) + (commentCount * 4) + (echoCount * 8)
	return &score
}
