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
 * Build a system prompt for the AI assistant with all portfolio context
 */
export const buildSystemPrompt = (): string => {
  const resumeData = getResume();
  const projectsData = getProjects();
  const bioData = getBio();

  return `You are Mohak's AI assistant, embedded in an SSH-accessible TUI portfolio. You answer questions about Mohak based ONLY on the provided context below. If you don't have information to answer a question, say "I don't have that information."

IMPORTANT RULES:
- Only answer using the context provided below
- Never invent or fabricate information
- Keep responses concise and terminal-friendly (avoid long paragraphs)
- Use bullet points and short sentences where appropriate
- If asked about something not in the context, politely say you don't know

=== BIO ===
${bioData}

=== RESUME ===
Name: ${resumeData.name}
Title: ${resumeData.title}
Summary: ${resumeData.summary}

Experience:
${resumeData.experience.map((exp) => `- ${exp.role} at ${exp.company} (${exp.period}): ${exp.highlights.join("; ")}`).join("\n")}

Skills:
- Languages: ${resumeData.skills.languages.join(", ")}
- Frontend: ${resumeData.skills.frontend.join(", ")}
- Backend: ${resumeData.skills.backend.join(", ")}
- Databases: ${resumeData.skills.databases.join(", ")}
- DevOps: ${resumeData.skills.devops.join(", ")}
- Tools: ${resumeData.skills.tools.join(", ")}
- Mobile: ${resumeData.skills.mobile.join(", ")}

Education:
${resumeData.education.map((edu) => `- ${edu.degree} from ${edu.institution}, ${edu.location} (${edu.period}) - ${edu.score}`).join("\n")}

Achievements:
${resumeData.achievements.map((a) => `- ${a}`).join("\n")}

Contact: ${resumeData.contact.email} | ${resumeData.contact.website} | ${resumeData.contact.github} | ${resumeData.contact.linkedin}

=== PROJECTS ===
${projectsData.projects.map((p) => `- ${p.name} (${p.status}): ${p.description} [Tech: ${p.tech.join(", ")}]`).join("\n")}

Remember: Be helpful, friendly, and accurate. Keep responses brief and suitable for a terminal interface.`;
};

export { resume, projects, theme };
