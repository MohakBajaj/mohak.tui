package ai

import (
	"fmt"
	"strings"

	"github.com/mohakbajaj/mohak-tui/apps/tui-server/internal/content"
)

// PromptBuilder builds system prompts using embedded portfolio content.
type PromptBuilder struct {
	resume   *content.Resume
	projects *content.Projects
	bio      string
}

// NewPromptBuilder creates a prompt builder from loaded content.
func NewPromptBuilder(resume *content.Resume, projects *content.Projects, bio string) *PromptBuilder {
	return &PromptBuilder{
		resume:   resume,
		projects: projects,
		bio:      bio,
	}
}

// BuildSystemPrompt returns a context-aware system prompt.
func (b *PromptBuilder) BuildSystemPrompt(userMessage string) string {
	intent := IntentGeneral
	if userMessage != "" {
		intent = DetectQueryIntent(userMessage)
	}

	context := b.buildContextForIntent(intent)
	return fmt.Sprintf(`You are NEURAL, Mohak's AI assistant embedded in an SSH-accessible TUI portfolio (ssh bmohak.xyz).

## PERSONA
You are helpful, concise, and technically knowledgeable. You have a subtle cyberpunk personality that matches the terminal aesthetic-professional but with character. Use technical language appropriately.

## CORE RULES
1. ONLY use information from the CONTEXT below - never invent details
2. Keep responses terminal-friendly:
   - Max 3-4 short paragraphs
   - Use bullet points for lists
   - Avoid walls of text
3. Be accurate: If information isn't in the context, say "I don't have that information about Mohak"
4. Be conversational: You can use first person ("Mohak is..." not "The user is...")
5. Formatting:
   - Use **bold** for emphasis
   - Use code formatting for technical terms
   - Use bullet points (•) for lists

## RESPONSE PATTERNS
- Greetings: Brief, friendly intro mentioning you're Mohak's AI assistant
- Technical questions: Be specific, mention exact technologies
- Experience questions: Highlight relevant roles and achievements
- Vague questions: Ask for clarification or provide overview

## WHAT NOT TO DO
- Don't make up work experience, projects, or skills
- Don't provide information about topics outside the context
- Don't write essays-keep it scannable
- Don't use emojis excessively
- Don't break character

---

## CONTEXT

%s

---

Remember: You represent Mohak's professional portfolio. Be helpful, accurate, and keep responses optimized for terminal display.`, context)
}

// GenerateFollowUps returns intent-aware suggestion prompts.
func GenerateFollowUps(intent QueryIntent) []string {
	followUps := map[QueryIntent][]string{
		IntentGreeting:     {"What are Mohak's main skills?", "Tell me about his experience", "What projects has he built?"},
		IntentAbout:        {"What's his work experience?", "What technologies does he use?", "How can I contact him?"},
		IntentExperience:   {"What technologies did he use?", "Tell me about his projects", "What are his achievements?"},
		IntentSkills:       {"What projects showcase these skills?", "Where has he applied these?", "What's his strongest area?"},
		IntentProjects:     {"What tech stack was used?", "Is it open source?", "Any live demos?"},
		IntentContact:      {"Tell me more about him", "What's his experience?", "See his projects"},
		IntentEducation:    {"What did he learn?", "Any achievements during college?", "What's his work experience?"},
		IntentAchievements: {"Tell me about his projects", "What's his experience?", "What skills does he have?"},
		IntentMeta:         {"What tech was used to build this?", "Tell me about Mohak", "See his other projects"},
		IntentGeneral:      {"What are his skills?", "See his projects", "How to contact him?"},
	}

	return followUps[intent]
}

func (b *PromptBuilder) buildContextForIntent(intent QueryIntent) string {
	sections := []string{
		fmt.Sprintf("# MOHAK BAJAJ\n%s\n\"%s\"\n\n%s", b.resume.Title, b.resume.Tagline, b.resume.Summary),
	}

	switch intent {
	case IntentGreeting, IntentAbout:
		sections = append(sections, "# BIO\n"+b.bio, buildContactSection(b.resume))
	case IntentExperience:
		sections = append(sections, buildExperienceSection(b.resume))
	case IntentSkills:
		sections = append(sections, buildSkillsSection(b.resume))
	case IntentProjects:
		sections = append(sections, buildProjectsSection(b.projects))
	case IntentContact:
		sections = append(sections, buildContactSection(b.resume))
	case IntentEducation:
		sections = append(sections, buildEducationSection(b.resume))
	case IntentAchievements:
		sections = append(sections, buildAchievementsSection(b.resume))
	case IntentMeta:
		if sshProject := b.projects.GetProjectByID("ssh-portfolio"); sshProject != nil {
			sections = append(sections, fmt.Sprintf(
				"# THIS APPLICATION\n%s: %s\nTech Stack: %s\nDemo: %s\nSource: %s",
				sshProject.Name,
				sshProject.Description,
				strings.Join(sshProject.Tech, ", "),
				sshProject.Links.Demo,
				sshProject.Links.Github,
			))
		}
		sections = append(sections, buildSkillsSection(b.resume))
	default:
		sections = append(
			sections,
			"# BIO\n"+b.bio,
			buildExperienceSection(b.resume),
			buildSkillsSection(b.resume),
			buildProjectsSection(b.projects),
			buildEducationSection(b.resume),
			buildAchievementsSection(b.resume),
			buildContactSection(b.resume),
		)
	}

	return strings.Join(sections, "\n\n")
}

func buildExperienceSection(resume *content.Resume) string {
	parts := make([]string, 0, len(resume.Experience))
	for _, experience := range resume.Experience {
		parts = append(parts, fmt.Sprintf(
			"## %s @ %s\n**%s**\n%s",
			experience.Role,
			experience.Company,
			experience.Period,
			bulletLines(experience.Highlights),
		))
	}

	return "# EXPERIENCE\n" + strings.Join(parts, "\n\n")
}

func buildSkillsSection(resume *content.Resume) string {
	return fmt.Sprintf(`# TECHNICAL SKILLS
• **Languages:** %s
• **Frontend:** %s
• **Backend:** %s
• **Databases:** %s
• **DevOps:** %s
• **Tools:** %s
• **Mobile:** %s`,
		strings.Join(resume.Skills.Languages, ", "),
		strings.Join(resume.Skills.Frontend, ", "),
		strings.Join(resume.Skills.Backend, ", "),
		strings.Join(resume.Skills.Databases, ", "),
		strings.Join(resume.Skills.DevOps, ", "),
		strings.Join(resume.Skills.Tools, ", "),
		strings.Join(resume.Skills.Mobile, ", "),
	)
}

func buildProjectsSection(projects *content.Projects) string {
	parts := make([]string, 0, len(projects.Projects))
	for _, project := range projects.Projects {
		lines := []string{
			fmt.Sprintf("## %s [%s]", project.Name, project.Status),
			project.Description,
			fmt.Sprintf("**Tech:** %s", strings.Join(project.Tech, ", ")),
		}
		if project.Links.Demo != "" {
			lines = append(lines, fmt.Sprintf("**Demo:** %s", project.Links.Demo))
		}
		if project.Links.Github != "" {
			lines = append(lines, fmt.Sprintf("**Source:** %s", project.Links.Github))
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	return "# PROJECTS\n" + strings.Join(parts, "\n\n")
}

func buildEducationSection(resume *content.Resume) string {
	parts := make([]string, 0, len(resume.Education))
	for _, education := range resume.Education {
		parts = append(parts, fmt.Sprintf(
			"• **%s** - %s, %s\n  %s | %s",
			education.Degree,
			education.Institution,
			education.Location,
			education.Period,
			education.Score,
		))
	}

	return "# EDUCATION\n" + strings.Join(parts, "\n")
}

func buildAchievementsSection(resume *content.Resume) string {
	return "# ACHIEVEMENTS\n" + bulletLines(resume.Achievements)
}

func buildContactSection(resume *content.Resume) string {
	return fmt.Sprintf(`# CONTACT
• **Email:** %s
• **Website:** %s
• **GitHub:** %s
• **LinkedIn:** %s
• **Twitter:** %s`,
		resume.Contact.Email,
		resume.Contact.Website,
		resume.Contact.Github,
		resume.Contact.LinkedIn,
		resume.Contact.Twitter,
	)
}

func bulletLines(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "• "+item)
	}
	return strings.Join(lines, "\n")
}
