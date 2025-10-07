package domain

import (
	"time"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type ContributionType string

const (
	ContribText  ContributionType = constants.MessageTypeText
	ContribVoice ContributionType = constants.MessageTypeVoice
	ContribCall  ContributionType = "call"
)

type Contribution struct {
	Type    ContributionType
	TextLen int // for text
	Seconds int // for voicenote/call
	At      time.Time
}

type ScoreSnapshot struct {
	Threshold int
	Me        int
	Them      int
	Revealed  bool
}

// pseudo: load latest config once per minute
type ScoreCfg struct {
	Threshold int
	// text
	TextBase, TextPerChar, TextMax float64
	TextCooldownSec                int
	// voice
	VoicePerSec, VoiceMax float64
	VoiceMinSec           int
	// call
	CallPerMin, CallMax float64
	CallMinSec          int
	// bonuses
	FirstMsgOfDay  float64
	ReplyWithinSec int
	ReplyBonus     float64
	// penalties
	DupWindowSec  int
	MaxMsgsPerMin int
}
