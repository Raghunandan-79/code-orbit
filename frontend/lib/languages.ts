import type { SupportedLanguage } from "@/lib/types";

export interface LanguageMeta {
  id: SupportedLanguage;
  label: string;
  monacoId: string;
  icon: string;       // emoji fallback; replace with SVG icons in production
  extension: string;
  color: string;      // brand accent for this language
}

export const LANGUAGES: LanguageMeta[] = [
  {
    id: "javascript",
    label: "JavaScript",
    monacoId: "javascript",
    icon: "JS",
    extension: ".js",
    color: "#f7df1e",
  },
  {
    id: "typescript",
    label: "TypeScript",
    monacoId: "typescript",
    icon: "TS",
    extension: ".ts",
    color: "#3178c6",
  },
  {
    id: "python",
    label: "Python",
    monacoId: "python",
    icon: "PY",
    extension: ".py",
    color: "#3572a5",
  },
  {
    id: "go",
    label: "Go",
    monacoId: "go",
    icon: "GO",
    extension: ".go",
    color: "#00add8",
  },
  {
    id: "rust",
    label: "Rust",
    monacoId: "rust",
    icon: "RS",
    extension: ".rs",
    color: "#ce412b",
  },
  {
    id: "java",
    label: "Java",
    monacoId: "java",
    icon: "JV",
    extension: ".java",
    color: "#b07219",
  },
  {
    id: "cpp",
    label: "C++",
    monacoId: "cpp",
    icon: "C+",
    extension: ".cpp",
    color: "#f34b7d",
  },
  {
    id: "c",
    label: "C",
    monacoId: "c",
    icon: "C",
    extension: ".c",
    color: "#555555",
  },
  {
    id: "csharp",
    label: "C#",
    monacoId: "csharp",
    icon: "C#",
    extension: ".cs",
    color: "#178600",
  },
];

export const LANGUAGE_MAP = new Map<SupportedLanguage, LanguageMeta>(
  LANGUAGES.map((l) => [l.id, l])
);

export function getLanguageMeta(id: SupportedLanguage): LanguageMeta {
  return LANGUAGE_MAP.get(id) ?? LANGUAGES[0];
}

/** Default starter code for each language */
export const LANGUAGE_STARTERS: Record<SupportedLanguage, string> = {
  javascript: `// Welcome to Code Orbit!\nconsole.log("Hello, world!");\n`,
  typescript: `// Welcome to Code Orbit!\nconst greet = (name: string): string => {\n  return \`Hello, \${name}!\`;\n};\n\nconsole.log(greet("world"));\n`,
  python: `# Welcome to Code Orbit!\ndef greet(name: str) -> str:\n    return f"Hello, {name}!"\n\nprint(greet("world"))\n`,
  go: `// Welcome to Code Orbit!\npackage main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello, world!")\n}\n`,
  rust: `// Welcome to Code Orbit!\nfn main() {\n    println!("Hello, world!");\n}\n`,
  java: `// Welcome to Code Orbit!\npublic class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello, world!");\n    }\n}\n`,
  cpp: `// Welcome to Code Orbit!\n#include <iostream>\n\nint main() {\n    std::cout << "Hello, world!" << std::endl;\n    return 0;\n}\n`,
  c: `// Welcome to Code Orbit!\n#include <stdio.h>\n\nint main() {\n    printf("Hello, world!\\n");\n    return 0;\n}\n`,
  csharp: `// Welcome to Code Orbit!\nusing System;\n\nclass Program {\n    static void Main() {\n        Console.WriteLine("Hello, world!");\n    }\n}\n`,
};
