import resume from "./resume.json";
import projects from "./projects.json";
import theme from "./theme.json";
import { readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));

export interface Resume {
  name: string;
  title: string;
  tagline: string;
  contact: {
    email: string;
    website: string;
    github: string;
    linkedin: string;
    twitter: string;
  };
  summary: string;
  experience: {
    company: string;
    role: string;
    period: string;
    highlights: string[];
  }[];
  skills: {
    languages: string[];
    frontend: string[];
    backend: string[];
    databases: string[];
    devops: string[];
    tools: string[];
    mobile: string[];
  };
  education: {
    institution: string;
    degree: string;
    location: string;
    period: string;
    score: string;
  }[];
  achievements: string[];
}

export interface Project {
  id: string;
  name: string;
  description: string;
  tech: string[];
  status: "active" | "completed" | "archived";
  links: {
    demo?: string;
    github?: string;
  };
}

export interface Projects {
  projects: Project[];
}

export interface ThemeColors {
  name: string;
  background: string;
  foreground: string;
  primary: string;
  secondary: string;
  accent: string;
  error: string;
  warning: string;
  success: string;
  muted: string;
  border: string;
  highlight: string;
}

export interface Theme {
  dark: ThemeColors;
  light: ThemeColors;
}

export const getResume = (): Resume => resume as Resume;
export const getProjects = (): Projects => projects as Projects;
export const getTheme = (): Theme => theme as Theme;

export const getBio = (): string => {
  return readFileSync(join(__dirname, "bio.md"), "utf-8");
};

/**
 * Query intent classification for smarter responses
 */
export type QueryIntent =
  | "greeting"
  | "about"
  | "experience"
  | "skills"
  | "projects"
  | "contact"
  | "education"
  | "achievements"
  | "general"
  | "meta"; // questions about the portfolio itself

/**
 * Detect query intent for context-aware responses
 */
export function detectQueryIntent(query: string): QueryIntent {
  const q = query.toLowerCase();

  // Greetings
  if (/^(hi|hello|hey|greetings|sup|yo)\b/.test(q)) return "greeting";

  // Meta questions about the portfolio
  if (
    /\b(this (app|portfolio|tui)|how (does|do) (this|you) work|built with|made with)\b/.test(
      q,
    )
  )
    return "meta";

  // About/Bio
  if (
    /\b(who (is|are)|about|bio|introduce|tell me about (mohak|yourself|him))\b/.test(
      q,
    )
  )
    return "about";

  // Experience/Work
  if (
    /\b(experience|work|job|company|role|position|career|employed|working)\b/.test(
      q,
    )
  )
    return "experience";

  // Skills
  if (
    /\b(skill|tech|technology|stack|language|framework|tool|know|proficient|expertise)\b/.test(
      q,
    )
  )
    return "skills";

  // Projects
  if (/\b(project|built|build|create|made|portfolio|app|application)\b/.test(q))
    return "projects";

  // Contact
  if (
    /\b(contact|reach|email|twitter|github|linkedin|social|hire|connect)\b/.test(
      q,
    )
  )
    return "contact";

  // Education
  if (
    /\b(education|school|college|university|degree|study|graduate|cgpa|gpa)\b/.test(
      q,
    )
  )
    return "education";

  // Achievements
  if (
    /\b(achievement|award|accomplish|win|competition|hackathon|volunteer)\b/.test(
      q,
    )
  )
    return "achievements";

  return "general";
}

/**
 * Build optimized context based on query intent
 */
function buildContextForIntent(intent: QueryIntent): string {
  const resumeData = getResume();
  const projectsData = getProjects();
  const bioData = getBio();

  const sections: string[] = [];

  // Always include basic identity
  sections.push(`# MOHAK BAJAJ
${resumeData.title}
"${resumeData.tagline}"

${resumeData.summary}`);

  // Add relevant sections based on intent
  switch (intent) {
    case "greeting":
    case "about":
      sections.push(`# BIO\n${bioData}`);
      sections.push(buildContactSection(resumeData));
      break;

    case "experience":
      sections.push(buildExperienceSection(resumeData));
      break;

    case "skills":
      sections.push(buildSkillsSection(resumeData));
      break;

    case "projects":
      sections.push(buildProjectsSection(projectsData));
      break;

    case "contact":
      sections.push(buildContactSection(resumeData));
      break;

    case "education":
      sections.push(buildEducationSection(resumeData));
      break;

    case "achievements":
      sections.push(buildAchievementsSection(resumeData));
      break;

    case "meta":
      // Include info about this SSH portfolio
      const sshProject = projectsData.projects.find(
        (p) => p.id === "ssh-portfolio",
      );
      if (sshProject) {
        sections.push(`# THIS APPLICATION
${sshProject.name}: ${sshProject.description}
Tech Stack: ${sshProject.tech.join(", ")}
Demo: ${sshProject.links.demo}
Source: ${sshProject.links.github}`);
      }
      sections.push(buildSkillsSection(resumeData));
      break;

    case "general":
    default:
      // Include everything for general queries
      sections.push(`# BIO\n${bioData}`);
      sections.push(buildExperienceSection(resumeData));
      sections.push(buildSkillsSection(resumeData));
      sections.push(buildProjectsSection(projectsData));
      sections.push(buildEducationSection(resumeData));
      sections.push(buildAchievementsSection(resumeData));
      sections.push(buildContactSection(resumeData));
      break;
  }

  return sections.join("\n\n");
}

function buildExperienceSection(r: Resume): string {
  return `# EXPERIENCE
${r.experience
  .map(
    (exp) => `## ${exp.role} @ ${exp.company}
**${exp.period}**
${exp.highlights.map((h) => `• ${h}`).join("\n")}`,
  )
  .join("\n\n")}`;
}

function buildSkillsSection(r: Resume): string {
  return `# TECHNICAL SKILLS
• **Languages:** ${r.skills.languages.join(", ")}
• **Frontend:** ${r.skills.frontend.join(", ")}
• **Backend:** ${r.skills.backend.join(", ")}
• **Databases:** ${r.skills.databases.join(", ")}
• **DevOps:** ${r.skills.devops.join(", ")}
• **Tools:** ${r.skills.tools.join(", ")}
• **Mobile:** ${r.skills.mobile.join(", ")}`;
}

function buildProjectsSection(p: Projects): string {
  return `# PROJECTS
${p.projects
  .map(
    (proj) => `## ${proj.name} [${proj.status}]
${proj.description}
**Tech:** ${proj.tech.join(", ")}
${proj.links.demo ? `**Demo:** ${proj.links.demo}` : ""}
${proj.links.github ? `**Source:** ${proj.links.github}` : ""}`,
  )
  .join("\n\n")}`;
}

function buildEducationSection(r: Resume): string {
  return `# EDUCATION
${r.education
  .map(
    (edu) => `• **${edu.degree}** - ${edu.institution}, ${edu.location}
  ${edu.period} | ${edu.score}`,
  )
  .join("\n")}`;
}

function buildAchievementsSection(r: Resume): string {
  return `# ACHIEVEMENTS
${r.achievements.map((a) => `• ${a}`).join("\n")}`;
}

function buildContactSection(r: Resume): string {
  return `# CONTACT
• **Email:** ${r.contact.email}
• **Website:** ${r.contact.website}
• **GitHub:** ${r.contact.github}
• **LinkedIn:** ${r.contact.linkedin}
• **Twitter:** ${r.contact.twitter}`;
}

/**
 * Build the full system prompt with persona and rules
 */
export function buildSystemPrompt(userMessage?: string): string {
  const intent = userMessage ? detectQueryIntent(userMessage) : "general";
  const context = buildContextForIntent(intent);

  return `You are NEURAL, Mohak's AI assistant embedded in an SSH-accessible TUI portfolio (ssh mohak.sh).

## PERSONA
You are helpful, concise, and technically knowledgeable. You have a subtle cyberpunk personality that matches the terminal aesthetic—professional but with character. Use technical language appropriately.

## CORE RULES
1. **ONLY use information from the CONTEXT below** - never invent details
2. **Keep responses terminal-friendly:**
   - Max 3-4 short paragraphs
   - Use bullet points for lists
   - Avoid walls of text
3. **Be accurate:** If information isn't in the context, say "I don't have that information about Mohak"
4. **Be conversational:** You can use first person ("Mohak is..." not "The user is...")
5. **Formatting:**
   - Use **bold** for emphasis
   - Use \`code\` for technical terms
   - Use bullet points (•) for lists

## RESPONSE PATTERNS
- **Greetings:** Brief, friendly intro mentioning you're Mohak's AI assistant
- **Technical questions:** Be specific, mention exact technologies
- **Experience questions:** Highlight relevant roles and achievements
- **Vague questions:** Ask for clarification or provide overview

## WHAT NOT TO DO
- Don't make up work experience, projects, or skills
- Don't provide information about topics outside the context
- Don't write essays—keep it scannable
- Don't use emojis excessively
- Don't break character

---

## CONTEXT

${context}

---

Remember: You represent Mohak's professional portfolio. Be helpful, accurate, and keep responses optimized for terminal display.`;
}

/**
 * Build a dynamic system prompt with intent-aware context
 */
export function buildDynamicSystemPrompt(
  userMessage: string,
  conversationHistory: { role: string; content: string }[] = [],
): string {
  // Analyze conversation for context
  const recentTopics = conversationHistory
    .slice(-4)
    .map((m) => detectQueryIntent(m.content));

  // Get primary intent from current message
  const primaryIntent = detectQueryIntent(userMessage);

  // Build optimized prompt
  return buildSystemPrompt(userMessage);
}

/**
 * Preprocess user message for better AI handling
 */
export function preprocessMessage(message: string): string {
  let processed = message.trim();

  // Normalize common variations
  processed = processed
    .replace(/\bu\b/gi, "you")
    .replace(/\bur\b/gi, "your")
    .replace(/\bpls\b/gi, "please")
    .replace(/\bthx\b/gi, "thanks");

  // Handle single-word queries
  if (processed.split(/\s+/).length === 1) {
    const word = processed.toLowerCase();
    if (
      ["skills", "experience", "projects", "contact", "education"].includes(
        word,
      )
    ) {
      processed = `Tell me about Mohak's ${word}`;
    }
  }

  return processed;
}

/**
 * Generate suggested follow-up questions based on response
 */
export function generateFollowUps(intent: QueryIntent): string[] {
  const followUps: Record<QueryIntent, string[]> = {
    greeting: [
      "What are Mohak's main skills?",
      "Tell me about his experience",
      "What projects has he built?",
    ],
    about: [
      "What's his work experience?",
      "What technologies does he use?",
      "How can I contact him?",
    ],
    experience: [
      "What technologies did he use?",
      "Tell me about his projects",
      "What are his achievements?",
    ],
    skills: [
      "What projects showcase these skills?",
      "Where has he applied these?",
      "What's his strongest area?",
    ],
    projects: [
      "What tech stack was used?",
      "Is it open source?",
      "Any live demos?",
    ],
    contact: [
      "Tell me more about him",
      "What's his experience?",
      "See his projects",
    ],
    education: [
      "What did he learn?",
      "Any achievements during college?",
      "What's his work experience?",
    ],
    achievements: [
      "Tell me about his projects",
      "What's his experience?",
      "What skills does he have?",
    ],
    meta: [
      "What tech was used to build this?",
      "Tell me about Mohak",
      "See his other projects",
    ],
    general: [
      "What are his skills?",
      "See his projects",
      "How to contact him?",
    ],
  };

  return followUps[intent] || followUps.general;
}

export { resume, projects, theme };
