package ai

import "regexp"

var (
	greetingPattern   = regexp.MustCompile(`^(hi|hello|hey|greetings|sup|yo)\b`)
	metaPattern       = regexp.MustCompile(`\b(this (app|portfolio|tui)|how (does|do) (this|you) work|built with|made with)\b`)
	aboutPattern      = regexp.MustCompile(`\b(who (is|are)|about|bio|introduce|tell me about (mohak|yourself|him))\b`)
	experiencePattern = regexp.MustCompile(
		`\b(experience|work|job|company|role|position|career|employed|working)\b`,
	)
	skillsPattern = regexp.MustCompile(
		`\b(skill|tech|technology|stack|language|framework|tool|know|proficient|expertise)\b`,
	)
	projectsPattern  = regexp.MustCompile(`\b(project|built|build|create|made|portfolio|app|application)\b`)
	contactPattern   = regexp.MustCompile(`\b(contact|reach|email|twitter|github|linkedin|social|hire|connect)\b`)
	educationPattern = regexp.MustCompile(
		`\b(education|school|college|university|degree|study|graduate|cgpa|gpa)\b`,
	)
	achievementsPattern = regexp.MustCompile(
		`\b(achievement|award|accomplish|win|competition|hackathon|volunteer)\b`,
	)
)

// DetectQueryIntent classifies a user query for context-aware prompting.
func DetectQueryIntent(query string) QueryIntent {
	q := toLower(query)

	switch {
	case greetingPattern.MatchString(q):
		return IntentGreeting
	case metaPattern.MatchString(q):
		return IntentMeta
	case aboutPattern.MatchString(q):
		return IntentAbout
	case experiencePattern.MatchString(q):
		return IntentExperience
	case skillsPattern.MatchString(q):
		return IntentSkills
	case projectsPattern.MatchString(q):
		return IntentProjects
	case contactPattern.MatchString(q):
		return IntentContact
	case educationPattern.MatchString(q):
		return IntentEducation
	case achievementsPattern.MatchString(q):
		return IntentAchievements
	default:
		return IntentGeneral
	}
}
