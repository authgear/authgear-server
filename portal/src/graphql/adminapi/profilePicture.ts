// Shared validation for the entry points that write
// user.standardAttributes.picture (ProfilePictureDialog and the file input on
// UserDetailSummary). Keep them in one place so they cannot drift apart on
// accepted type or size.

const MAX_FILE_SIZE = 500 * 1024;

const ACCEPTED_IMAGE_TYPES = new Set([
  "image/svg+xml",
  "image/png",
  "image/jpeg",
  "image/x-icon",
  "image/vnd.microsoft.icon",
]);

export const PROFILE_PICTURE_ACCEPT =
  ".svg,.png,.jpeg,.jpg,.ico,image/svg+xml,image/png,image/jpeg,image/x-icon,image/vnd.microsoft.icon";

export function isAcceptedProfilePictureFile(file: File): boolean {
  return (
    file.size <= MAX_FILE_SIZE &&
    (ACCEPTED_IMAGE_TYPES.has(file.type) ||
      /\.(svg|png|jpe?g|ico)$/i.test(file.name))
  );
}
