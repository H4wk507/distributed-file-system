export type FileType =
  | "image"
  | "video"
  | "audio"
  | "pdf"
  | "archive"
  | "doc"
  | "spreadsheet"
  | "text"
  | "file";

export function getFileIcon(contentType: string): FileType {
  if (contentType.startsWith("image/")) return "image";
  if (contentType.startsWith("video/")) return "video";
  if (contentType.startsWith("audio/")) return "audio";
  if (contentType === "application/pdf") return "pdf";
  if (
    contentType.includes("zip") ||
    contentType.includes("tar") ||
    contentType.includes("rar")
  )
    return "archive";
  if (
    contentType.includes("word") ||
    contentType.includes("document") ||
    contentType === "application/msword"
  )
    return "doc";
  if (contentType.includes("sheet") || contentType.includes("excel"))
    return "spreadsheet";
  if (contentType.startsWith("text/")) return "text";
  return "file";
}
