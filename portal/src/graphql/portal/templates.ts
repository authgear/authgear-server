import {
  LanguageTag,
  ResourceUpdate,
  Resource,
  validateLocales,
  diffResourceUpdates,
} from "../../util/resource";

export interface GenerateUpdatesResult {
  isModified: boolean;
  additions: ResourceUpdate[];
  invalidAdditionLocales: LanguageTag[];
  editions: ResourceUpdate[];
  invalidEditionLocales: LanguageTag[];
  deletions: ResourceUpdate[];
}

export function generateUpdates(
  initialTemplateLocales: LanguageTag[],
  initialTemplates: Resource[],
  templateLocales: LanguageTag[],
  templates: Resource[]
): GenerateUpdatesResult {
  const realTemplates = templates.filter(
    (t) => !t.specifier.locale || templateLocales.includes(t.specifier.locale)
  );
  const { isValid, invalidNewLocales, invalidDeletedLocales } = validateLocales(
    initialTemplateLocales,
    templateLocales,
    realTemplates
  );
  const {
    needUpdate,
    newResources,
    editedResources,
    deletedResources,
  } = diffResourceUpdates(initialTemplates, realTemplates);

  return {
    isModified: needUpdate || !isValid,
    additions: newResources,
    invalidAdditionLocales: invalidNewLocales,
    editions: editedResources,
    invalidEditionLocales: invalidDeletedLocales,
    deletions: deletedResources,
  };
}
