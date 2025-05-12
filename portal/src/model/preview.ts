export interface PreviewCustomisationMessage {
  type: "PreviewCustomisationMessage";
  theme: string;
  cssVars: Record<string, string>;
  images: Record<string, string | null>;
  translations: Record<string, string>;
  data: Record<string, string>;
}
