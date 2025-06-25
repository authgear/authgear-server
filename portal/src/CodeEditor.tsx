import React, { useCallback, useRef } from "react";
import ControlledEditor, {
  EditorProps,
  Monaco,
  OnChange,
  OnMount,
} from "@monaco-editor/react";

export interface CodeEditorProps extends EditorProps {
  className?: string;
  packages?: string[];
}

const DEFAULT_WHITELISTED_PACKAGES = [
  "https://deno.land/std",
  "https://deno.land/x/authgear_deno_hook",
];

const REGEX_DETECT_IMPORT =
  /(?:^|\r|\n|\r\n)(?:(?:(?:import)|(?:export))(?:.|\n)*?from\s+["']([^"']+)["'])|(?:require(?:\s+)?\(["']([^"']+)["']\))|(?:\/+\s+<reference\s+path=["']([^"']+)["']\s+\/>)/g;

function parseDependencies(source: string): string[] {
  return [...source.matchAll(REGEX_DETECT_IMPORT)]
    .map((x) => x[1] || x[2] || x[3])
    .filter((x) => !!x);
}

// Convert string like "https://deno.land/std" to a string safe for filename.
async function generateSafeFilename(pkg: string): Promise<string> {
  const utf8Bytes = new TextEncoder().encode(pkg);
  const sha256Bytes = await window.crypto.subtle.digest("SHA-256", utf8Bytes);
  const fileName = Array.from(new Uint8Array(sha256Bytes), (byte) =>
    byte.toString(16).padStart(2, "0")
  ).join("");
  return fileName;
}

const CodeEditor: React.VFC<CodeEditorProps> = function CodeEditor(props) {
  const {
    className,
    packages: whitelistedPackages = DEFAULT_WHITELISTED_PACKAGES,
    onMount,
    onChange,
    ...rest
  } = props;

  const isWhitelistedPackage = useCallback(
    (pkg: string) => {
      return whitelistedPackages.some(
        (allowedPkg) =>
          pkg === allowedPkg ||
          pkg.startsWith(allowedPkg + "@") || // versioned
          pkg.startsWith(allowedPkg + "/") // latest
      );
    },
    [whitelistedPackages]
  );

  const monacoRef = useRef<Monaco>();
  const resolveImports = useCallback(
    async (value: string | undefined) => {
      if (!value) {
        return;
      }

      const monaco = monacoRef.current;
      if (!monaco) {
        return;
      }

      // We only resolve first-level imports to avoid performance issue
      for (const pkg of parseDependencies(value)) {
        const fileName = await generateSafeFilename(pkg);
        const path = `inmemory://model/${fileName}.d.ts`;
        const uri = monaco.Uri.file(path);

        if (!monaco.editor.getModel(uri)) {
          // Here we only use best-effort to import external script using `declare module` due to following limitations:
          // - `monaco-editor` cannot use vscode extensions e.g. "vscode-deno", so cannot use official solution for url imports
          // -`--moduleResolution: bundle` in Typescript allows `.ts` extension but not yet supported in monaco-editor, thus path alias won't work.
          const code = isWhitelistedPackage(pkg)
            ? await fetch(pkg).then(async (r) => r.text())
            : "";
          const dts =
            `declare module '${pkg}'` +
            (code && parseDependencies(code).length === 0 ? `{${code}}` : "");
          monaco.languages.typescript.typescriptDefaults.addExtraLib(dts, path);
          monaco.editor.createModel(dts, "typescript", uri);
        }
      }
    },
    [isWhitelistedPackage]
  );

  const handleEditorMount = useCallback<OnMount>(
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    async (editor, monaco) => {
      monacoRef.current = monaco;

      const options =
        monaco.languages.typescript.typescriptDefaults.getCompilerOptions();
      options.strict = true;
      options.strictNullChecks = true;
      options.strictBindCallApply = true;
      options.moduleResolution =
        monaco.languages.typescript.ModuleResolutionKind.NodeJs;
      monaco.languages.typescript.typescriptDefaults.setCompilerOptions(
        options
      );

      // eslint-disable-next-line no-void
      void resolveImports(editor.getValue());

      onMount?.(editor, monaco);
    },
    [resolveImports, onMount]
  );

  const handleEditorChange = useCallback<OnChange>(
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    async (value, ev) => {
      // eslint-disable-next-line no-void
      void resolveImports(value);

      onChange?.(value, ev);
    },
    [onChange, resolveImports]
  );

  return (
    <div className={className}>
      <ControlledEditor
        height="100%"
        onMount={handleEditorMount}
        onChange={handleEditorChange}
        {...rest}
      />
    </div>
  );
};

export default CodeEditor;
