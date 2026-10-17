"use client";

import { useCallback, useEffect, useRef } from "react";
import dynamic from "next/dynamic";
import type * as Monaco from "monaco-editor";
import type { CollaborationState, CollaborationActions } from "@/hooks/useCollaboration";
import type { SupportedLanguage } from "@/lib/types";
import { getLanguageMeta, LANGUAGE_STARTERS } from "@/lib/languages";

// Monaco must be loaded client-side only
const MonacoEditor = dynamic(
  () => import("@monaco-editor/react").then((m) => m.default),
  {
    ssr: false,
    loading: () => <EditorSkeleton />,
  }
);

const CURSOR_COLORS = [
  "#6d6aff", "#34d399", "#fbbf24", "#f87171",
  "#60a5fa", "#f472b6", "#a78bfa", "#fb923c",
];

interface EditorPaneProps {
  language: SupportedLanguage | null;
  collab: CollaborationState & CollaborationActions;
  pendingUpdateRef: React.MutableRefObject<Uint8Array | null>;
}

export function EditorPane({ language, collab, pendingUpdateRef }: EditorPaneProps) {
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<typeof Monaco | null>(null);
  const decorationsRef = useRef<string[]>([]);
  const isApplyingRef = useRef(false);

  const langMeta = language ? getLanguageMeta(language) : null;
  const monacoLang = langMeta?.monacoId ?? "javascript";
  const starterCode = language ? LANGUAGE_STARTERS[language] : LANGUAGE_STARTERS.javascript;

  // Handle editor mount
  const handleEditorDidMount = useCallback(
    (editor: Monaco.editor.IStandaloneCodeEditor, monacoInstance: typeof Monaco) => {
      editorRef.current = editor;
      monacoRef.current = monacoInstance;

      // Set initial content
      const model = editor.getModel();
      if (model && !model.getValue()) {
        model.setValue(starterCode);
      }

      // Track cursor/selection for awareness
      editor.onDidChangeCursorPosition((e) => {
        collab.sendAwareness({
          nickname: "", // server knows our nickname from join
          line: e.position.lineNumber,
          column: e.position.column,
        });
      });

      editor.onDidChangeCursorSelection((e) => {
        const sel = e.selection;
        if (!sel.isEmpty()) {
          collab.sendAwareness({
            nickname: "",
            line: sel.startLineNumber,
            column: sel.startColumn,
            selection: {
              startLine: sel.startLineNumber,
              startColumn: sel.startColumn,
              endLine: sel.endLineNumber,
              endColumn: sel.endColumn,
            },
          });
        }
      });
    },
    [collab, starterCode]
  );

  // Apply remote content updates
  useEffect(() => {
    const interval = setInterval(() => {
      const update = pendingUpdateRef.current;
      if (!update || !editorRef.current) return;
      pendingUpdateRef.current = null;
      isApplyingRef.current = true;

      try {
        // Decode as UTF-8 text for simple text-based sync
        const text = new TextDecoder().decode(update);
        if (text && text !== editorRef.current.getValue()) {
          const model = editorRef.current.getModel();
          if (model) {
            const pos = editorRef.current.getPosition();
            model.pushEditOperations(
              [],
              [{ range: model.getFullModelRange(), text }],
              () => null
            );
            if (pos) editorRef.current.setPosition(pos);
          }
        }
      } catch {
        // binary update: ignore for now (CRDT integration point)
      } finally {
        isApplyingRef.current = false;
      }
    }, 50);
    return () => clearInterval(interval);
  }, [pendingUpdateRef]);

  // Send content on change
  const handleEditorChange = useCallback(
    (value: string | undefined) => {
      if (!value || isApplyingRef.current) return;
      // Encode as UTF-8 bytes and send
      const bytes = new TextEncoder().encode(value);
      collab.sendContentUpdate(bytes);
    },
    [collab]
  );

  // Draw remote cursors as Monaco decorations
  useEffect(() => {
    const editor = editorRef.current;
    if (!editor) return;

    const newDecorations: Monaco.editor.IModelDeltaDecoration[] = [];

    for (const cursor of collab.cursors) {
      const line = Math.max(1, cursor.line);
      const col = Math.max(1, cursor.column);

      // Cursor line decoration
      newDecorations.push({
        range: { startLineNumber: line, startColumn: col, endLineNumber: line, endColumn: col + 1 },
        options: {
          className: `remote-cursor-${cursor.colorIndex}`,
          beforeContentClassName: `remote-cursor-head-${cursor.colorIndex}`,
          stickiness: 1,
        },
      });

      // Selection decoration
      if (cursor.selection) {
        const s = cursor.selection;
        newDecorations.push({
          range: {
            startLineNumber: s.startLine,
            startColumn: s.startColumn,
            endLineNumber: s.endLine,
            endColumn: s.endColumn,
          },
          options: {
            className: `remote-selection-${cursor.colorIndex}`,
            stickiness: 1,
          },
        });
      }
    }

    decorationsRef.current = editor.deltaDecorations(
      decorationsRef.current,
      newDecorations
    );
  }, [collab.cursors]);

  // Inject dynamic cursor CSS
  useEffect(() => {
    const styleId = "remote-cursor-styles";
    let style = document.getElementById(styleId) as HTMLStyleElement | null;
    if (!style) {
      style = document.createElement("style");
      style.id = styleId;
      document.head.appendChild(style);
    }
    style.textContent = CURSOR_COLORS.map(
      (color, i) => `
        .remote-cursor-head-${i}::before {
          content: '';
          display: inline-block;
          width: 2px;
          height: 1.2em;
          background: ${color};
          margin-right: -2px;
          vertical-align: text-bottom;
          box-shadow: 0 0 8px ${color}90;
          border-radius: 1px;
        }
        .remote-selection-${i} {
          background: ${color}22;
          border-radius: 2px;
        }
      `
    ).join("\n");
    return () => style?.remove();
  }, []);

  return (
    <div className="w-full h-full flex flex-col bg-[#0d0d11]">
      {/* Tab bar */}
      <div
        className="shrink-0 flex items-center h-9 px-3 gap-1 border-b overflow-x-auto no-scrollbar"
        style={{
          background: "rgba(12,12,16,0.95)",
          borderColor: "rgba(255,255,255,0.06)",
        }}
      >
        <div
          className="flex items-center gap-2 px-3.5 h-full border-t-2 select-none"
          style={{
            borderTopColor: langMeta?.color ?? "var(--brand)",
            background: "rgba(255,255,255,0.04)",
            borderRight: "1px solid rgba(255,255,255,0.04)",
          }}
        >
          <span
            className="w-4 h-4 rounded flex items-center justify-center text-[9px] font-mono font-bold"
            style={{
              background: `${langMeta?.color ?? "#6366f1"}20`,
              color: langMeta?.color ?? "#6366f1",
            }}
          >
            {langMeta?.icon ?? "JS"}
          </span>
          <span className="text-xs font-mono text-[--fg-primary] whitespace-nowrap">
            main{langMeta?.extension ?? ".js"}
          </span>
        </div>
      </div>

      {/* Monaco */}
      <div className="flex-1 overflow-hidden">
        <MonacoEditor
          height="100%"
          language={monacoLang}
          theme="vs-dark"
          defaultValue={starterCode}
          onChange={handleEditorChange}
          onMount={handleEditorDidMount}
          options={{
            fontSize: 13,
            fontFamily: "var(--font-geist-mono), 'JetBrains Mono', 'Fira Code', monospace",
            fontLigatures: true,
            lineHeight: 22,
            letterSpacing: 0.3,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            wordWrap: "off",
            smoothScrolling: true,
            cursorBlinking: "smooth",
            cursorSmoothCaretAnimation: "on",
            renderLineHighlight: "gutter",
            renderWhitespace: "selection",
            guides: { bracketPairs: true, indentation: true },
            bracketPairColorization: { enabled: true },
            padding: { top: 12, bottom: 12 },
            suggest: { showIcons: true },
            formatOnPaste: true,
            formatOnType: false,
            accessibilitySupport: "auto",
            tabSize: 2,
            insertSpaces: true,
            automaticLayout: true,
            scrollbar: {
              verticalScrollbarSize: 6,
              horizontalScrollbarSize: 6,
              useShadows: false,
            },
          }}
        />
      </div>
    </div>
  );
}

function EditorSkeleton() {
  return (
    <div
      className="w-full h-full flex items-center justify-center bg-[#0d0d11]"
      aria-label="Loading editor…"
    >
      <div className="flex flex-col items-center gap-3">
        <div
          className="w-8 h-8 rounded-full border-2 border-t-[--brand] border-white/10 animate-spin"
          aria-hidden="true"
        />
        <span className="text-xs text-[--fg-tertiary]">Loading editor environment…</span>
      </div>
    </div>
  );
}
