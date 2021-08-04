export interface DownloadStringAsFileOptions {
  content: string;
  mimeType: string;
  filename: string;
}

export function downloadStringAsFile(
  options: DownloadStringAsFileOptions
): void {
  const { content, mimeType, filename } = options;

  const anchor = window.document.createElement("a");
  anchor.href = `data:${mimeType};charset=utf-8,${encodeURIComponent(content)}`;
  anchor.download = filename;
  anchor.style.display = "none";

  window.document.body.appendChild(anchor);
  anchor.click();
  window.document.body.removeChild(anchor);
}
