export function semanticToRadixColor(
  semantic: "error" | "success" | "info" | "warning"
): "red" | "green" | "sky" | "amber" {
  switch (semantic) {
    case "error":
      return "red";
    case "success":
      return "green";
    case "info":
      return "sky";
    case "warning":
      return "amber";
  }
}
