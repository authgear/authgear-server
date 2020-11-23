import { AppResourceUpdate } from "./__generated__/globalTypes";
import { TemplateLocale } from "../../templates";

export interface Template {
  locale: TemplateLocale;
  path: string;
  value: string;
}

export interface GenerateUpdatesResult {
  isModified: boolean;
  additions: AppResourceUpdate[];
  invalidAdditionLocales: TemplateLocale[];
  editions: AppResourceUpdate[];
  invalidEditionLocales: TemplateLocale[];
  deletions: AppResourceUpdate[];
}

// eslint-disable-next-line complexity
export function generateUpdates(
  initialTemplateLocales: TemplateLocale[],
  initialTemplates: Record<string, Template | undefined>,
  templateLocales: TemplateLocale[],
  templates: Record<string, Template | undefined>
): GenerateUpdatesResult {
  // We have 3 kinds of updates
  // 1. Addition
  // 2. Edition
  // 3. Deletion

  // Addition: present in templateLocales but absent in initialTemplateLocales
  const additionLocales: TemplateLocale[] = [];
  for (const locale of templateLocales) {
    const idx = initialTemplateLocales.indexOf(locale);
    if (idx < 0) {
      additionLocales.push(locale);
    }
  }
  // It is valid iff there is at least 1 template with non-empty value.
  const invalidAdditionLocales: TemplateLocale[] = [];
  const additions: AppResourceUpdate[] = [];
  for (const locale of additionLocales) {
    let valid = false;
    for (const template of Object.values(templates)) {
      if (template == null) {
        continue;
      }
      if (template.locale === locale && template.value !== "") {
        valid = true;
        additions.push({
          path: template.path,
          data: template.value,
        });
      }
    }
    if (!valid) {
      invalidAdditionLocales.push(locale);
    }
  }

  // Edition: present in both templateLocales and initialTemplateLocales
  const editionLocales: TemplateLocale[] = [];
  for (const locale of templateLocales) {
    const idx = initialTemplateLocales.indexOf(locale);
    if (idx >= 0) {
      editionLocales.push(locale);
    }
  }
  // It is valid iff there is at least 1 template with non-empty value.
  const invalidEditionLocales: TemplateLocale[] = [];
  const editions: AppResourceUpdate[] = [];
  for (const locale of editionLocales) {
    let valid = false;

    for (const template of Object.values(templates)) {
      if (template == null) {
        continue;
      }
      if (template.locale === locale) {
        if (template.value !== "") {
          valid = true;
        }

        for (const oldTemplate of Object.values(initialTemplates)) {
          if (oldTemplate == null) {
            continue;
          }
          if (
            oldTemplate.locale === template.locale &&
            oldTemplate.path === template.path
          ) {
            if (oldTemplate.value !== template.value) {
              editions.push({
                path: template.path,
                data: template.value === "" ? null : template.value,
              });
            }
          }
        }
      }
    }
    if (!valid) {
      invalidEditionLocales.push(locale);
    }
  }

  // Deletion: present in initialTemplateLocales but absent in templateLocales
  const deletionLocales: TemplateLocale[] = [];
  for (const locale of initialTemplateLocales) {
    const idx = templateLocales.indexOf(locale);
    if (idx < 0) {
      deletionLocales.push(locale);
    }
  }
  const deletions: AppResourceUpdate[] = [];
  for (const locale of deletionLocales) {
    for (const template of Object.values(initialTemplates)) {
      if (template == null) {
        continue;
      }
      if (template.locale === locale) {
        deletions.push({
          path: template.path,
          data: null,
        });
      }
    }
  }

  const isModified =
    additions.length > 0 ||
    invalidAdditionLocales.length > 0 ||
    editions.length > 0 ||
    invalidEditionLocales.length > 0 ||
    deletions.length > 0;

  return {
    isModified,
    additions,
    invalidAdditionLocales,
    editions,
    invalidEditionLocales,
    deletions,
  };
}
