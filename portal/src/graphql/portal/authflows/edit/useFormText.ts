import { useCallback, useMemo, useState } from "react";
import { useLiveState } from "../../../../hook/useSyncFormStates";
import { RESOURCE_TRANSLATION_JSON } from "../../../../resources";
import {
  ResourceSpecifier,
  decodeForPrettifiedJSON,
  expandSpecifier,
} from "../../../../util/resource";
import { escapeMessageFormatText } from "../../../../util/messageFormat";
import { useAppTemplatesQuery } from "../../query/appTemplatesQuery";
import { useUpdateAppTemplatesMutation } from "../../mutations/updateAppTemplatesMutation";
import {
  FormNameOp,
  TextEdits,
  filesWithRenames,
  hasFormKeys,
  planTextWrites,
  readEffectiveText,
  readOwnText,
  renameEdits,
  subtractEdits,
  withoutEdit,
} from "./formText";

type JSONFile = Record<string, unknown>;

export interface FormText {
  isLoading: boolean;
  loadError: unknown;
  updateError: unknown;
  isUpdating: boolean;
  // The project's own value, as stored (MessageFormat-escaped), with pending
  // edits applied; undefined when unset.
  ownText: (language: string, key: string) => string | undefined;
  // What the end-user sees now, as stored: the own value, else the built-in.
  effectiveText: (language: string, key: string) => string | undefined;
  // The saved value, renames applied; undefined when unset. Whether the
  // builder may edit a text is decided from it, not from what is typed.
  savedText: (language: string, key: string) => string | undefined;
  // Sets a key from plain text; null removes it.
  setText: (language: string, key: string, value: string | null) => void;
  // Drops a pending edit, so the saved value shows again.
  clearEdit: (language: string, key: string) => void;
  recordNameOp: (op: FormNameOp) => void;
  // Whether the form has any text of its own in the language.
  hasText: (language: string, formName: string) => boolean;
  // The project's saved files, by language; null when a language has none.
  files: Record<string, JSONFile | null>;
  edits: TextEdits;
  ops: FormNameOp[];
  restore: (edits: TextEdits, ops: FormNameOp[]) => void;
  getIsDirty: () => boolean;
  reset: () => void;
  // Writes the pending text onto the latest saved files.
  save: (usedAfter: Set<string>) => Promise<void>;
}

function parseJSON(value: string | null | undefined): JSONFile | null {
  if (value == null || value === "") {
    return null;
  }
  try {
    const parsed: unknown = JSON.parse(value);
    return typeof parsed === "object" &&
      parsed != null &&
      !Array.isArray(parsed)
      ? (parsed as JSONFile)
      : null;
  } catch {
    return null;
  }
}

function specifierOf(language: string): ResourceSpecifier {
  return { def: RESOURCE_TRANSLATION_JSON, locale: language, extension: null };
}

// translation.json of every language, and the form text edited in the
// builder but not saved yet.
export function useFormText(appID: string, languages: string[]): FormText {
  const specifiers = useMemo(() => languages.map(specifierOf), [languages]);
  const {
    resources,
    loading: isLoading,
    error: loadError,
    refetch,
  } = useAppTemplatesQuery(appID, specifiers);
  const {
    updateAppTemplates,
    error: mutationError,
    resetError,
  } = useUpdateAppTemplatesMutation(appID);
  // Covers the whole save, refetches included, so nothing is edited while
  // it runs and a failed refetch is shown, not dropped.
  const [isUpdating, setIsUpdating] = useState(false);
  const [saveError, setSaveError] = useState<unknown>(null);
  const updateError: unknown = saveError ?? mutationError;

  const { files, effective } = useMemo(() => {
    const files: Record<string, JSONFile | null> = {};
    const effective: Record<string, JSONFile> = {};
    for (const r of resources) {
      const language = r.specifier.locale ?? "";
      // The query fills an absent file with "{}"; an empty own file and an
      // absent one are written the same way.
      files[language] = parseJSON(r.nullableValue);
      effective[language] = parseJSON(r.effectiveData) ?? {};
    }
    return { files, effective };
  }, [resources]);

  // getPending reads the latest edits outside render, e.g. right after a
  // save resolves.
  const [pending, setPending, getPending] = useLiveState<{
    edits: TextEdits;
    ops: FormNameOp[];
  }>({ edits: {}, ops: [] });
  const { edits, ops } = pending;

  // A renamed form shows its text under the new name before Save.
  const shownFiles = useMemo(() => filesWithRenames(files, ops), [files, ops]);

  const ownText = useCallback(
    (language: string, key: string) =>
      readOwnText(edits, shownFiles, language, key),
    [edits, shownFiles]
  );

  const effectiveText = useCallback(
    (language: string, key: string) =>
      readEffectiveText(edits, shownFiles, effective, language, key),
    [edits, shownFiles, effective]
  );

  const savedText = useCallback(
    (language: string, key: string) =>
      readOwnText({}, shownFiles, language, key),
    [shownFiles]
  );

  const setText = useCallback(
    (language: string, key: string, value: string | null) => {
      setPending((prev) => ({
        ...prev,
        edits: {
          ...prev.edits,
          [language]: {
            ...prev.edits[language],
            [key]: value == null ? null : escapeMessageFormatText(value),
          },
        },
      }));
    },
    [setPending]
  );

  const clearEdit = useCallback(
    (language: string, key: string) => {
      setPending((prev) => ({
        ...prev,
        edits: withoutEdit(prev.edits, language, key),
      }));
    },
    [setPending]
  );

  const recordNameOp = useCallback(
    (op: FormNameOp) => {
      setPending((prev) => ({
        // Text typed under the old name moves with the form.
        edits:
          op.kind === "rename"
            ? renameEdits(prev.edits, op.from, op.to)
            : prev.edits,
        ops: [...prev.ops, op],
      }));
    },
    [setPending]
  );

  const hasText = useCallback(
    (language: string, formName: string): boolean => {
      const file: JSONFile = { ...shownFiles[language] };
      for (const [key, value] of Object.entries(edits[language] ?? {})) {
        if (value == null) {
          delete file[key];
        } else {
          file[key] = value;
        }
      }
      return hasFormKeys(file, formName);
    },
    [shownFiles, edits]
  );

  const restore = useCallback(
    (e: TextEdits, o: FormNameOp[]) => {
      setPending({ edits: e, ops: o });
    },
    [setPending]
  );

  const getIsDirty = useCallback(() => {
    const p = getPending();
    return (
      p.ops.length > 0 ||
      Object.values(p.edits).some((e) => Object.keys(e).length > 0)
    );
  }, [getPending]);

  const reset = useCallback(() => {
    resetError();
    setSaveError(null);
    setPending({ edits: {}, ops: [] });
  }, [resetError, setPending]);

  const save = useCallback(
    async (usedAfter: Set<string>) => {
      const snapshot = getPending();
      if (!getIsDirty()) {
        return;
      }
      setIsUpdating(true);
      setSaveError(null);
      try {
        // The latest files, so text changed elsewhere since loading is kept.
        const result = await refetch();
        const node =
          result.data.node?.__typename === "App" ? result.data.node : null;
        const latest: Record<string, JSONFile | null> = {};
        const checksums: Record<string, string | null | undefined> = {};
        for (const language of languages) {
          const path = expandSpecifier(specifierOf(language));
          const resource = node?.resources.find((r) => r.path === path);
          latest[language] =
            resource?.data != null
              ? parseJSON(decodeForPrettifiedJSON(resource.data))
              : null;
          checksums[language] = resource?.checksum;
        }
        const writes = planTextWrites(
          latest,
          snapshot.edits,
          snapshot.ops,
          usedAfter
        );
        const updates = Object.entries(writes).map(([language, file]) => {
          const specifier = specifierOf(language);
          return {
            specifier,
            path: expandSpecifier(specifier),
            value: JSON.stringify(file),
            checksum: checksums[language],
          };
        });
        if (updates.length > 0) {
          await updateAppTemplates(updates);
          // Shows the saved files once the edits are cleared.
          await refetch();
        }
        // Only what was saved; text typed meanwhile stays pending.
        setPending((prev) => ({
          edits: subtractEdits(prev.edits, snapshot.edits),
          ops: prev.ops.slice(snapshot.ops.length),
        }));
      } catch (e: unknown) {
        setSaveError(e);
        throw e;
      } finally {
        setIsUpdating(false);
      }
    },
    [getPending, getIsDirty, refetch, languages, updateAppTemplates, setPending]
  );

  return {
    isLoading,
    loadError,
    updateError,
    isUpdating,
    ownText,
    effectiveText,
    savedText,
    setText,
    clearEdit,
    recordNameOp,
    hasText,
    files,
    edits,
    ops,
    restore,
    getIsDirty,
    reset,
    save,
  };
}
