package domain

type Steps string

const (
	OnboardingStepsUnset             Steps = "UNSET"
	OnboardingStepsNone              Steps = "NONE"
	OnboardingStepsIntro             Steps = "INTRO"
	OnboardingStepsBasics            Steps = "BASICS"
	OnboardingStepsLocation          Steps = "LOCATION"
	OnboardingStepsLifestyle         Steps = "LIFESTYLE"
	OnboardingStepsBeliefs           Steps = "BELIEFS"
	OnboardingStepsBackground        Steps = "BACKGROUND"
	OnboardingStepsWorkAndEducation  Steps = "WORK_AND_EDUCATION"
	OnboardingStepsLanguages         Steps = "LANGUAGES"
	OnboardingStepsPhotos            Steps = "PHOTOS"
	OnboardingStepsPrompts           Steps = "PROMPTS"
	OnboardingStepsProfile           Steps = "PROFILE"
	OnboardingStepsVideoVerification Steps = "VIDEO_VERIFICATION"
	OnboardingStepsComplete          Steps = "COMPLETE"
)

// GenerateOnboardingSteps generates the OnboardingSteps result/response to return to the frontend for step s.
func (s Steps) GenerateOnboardingSteps() OnboardingSteps {
	var onboardingSteps OnboardingSteps
	onboardingSteps.TotalSteps = len(OrderedSteps)
	onboardingSteps.Steps = OrderedSteps

	currentIdx := 0

	for i, step := range OrderedSteps {
		if step == s {
			// prev
			if i-1 < 0 {
				onboardingSteps.PreviousStep = OnboardingStepsNone
			} else {
				onboardingSteps.PreviousStep = OrderedSteps[i-1]
			}

			// current
			onboardingSteps.CurrentStep = s

			// next (guard end)
			if i+1 < len(OrderedSteps) {
				onboardingSteps.NextStep = OrderedSteps[i+1]
			} else {
				onboardingSteps.NextStep = OnboardingStepsNone
			}

			currentIdx = i + 1 // 1-based position

			break
		}
	}

	// percentage with float division
	onboardingSteps.Progress = (float64(currentIdx) / float64(len(OrderedSteps))) * 100.0

	return onboardingSteps
}

func (s Steps) NextStep() Steps {
	for i, step := range OrderedSteps {
		if step == s {
			if i+1 < len(OrderedSteps) {
				return OrderedSteps[i+1]
			} else {
				return OnboardingStepsNone
			}
		}
	}

	return OnboardingStepsNone
}

var OrderedSteps = []Steps{
	OnboardingStepsIntro,
	OnboardingStepsBasics,
	OnboardingStepsLocation,
	OnboardingStepsLifestyle,
	OnboardingStepsBeliefs,
	OnboardingStepsBackground,
	OnboardingStepsWorkAndEducation,
	OnboardingStepsLanguages,
	OnboardingStepsPhotos,
	OnboardingStepsPrompts,
	OnboardingStepsProfile,
	OnboardingStepsVideoVerification,
	OnboardingStepsComplete,
}
