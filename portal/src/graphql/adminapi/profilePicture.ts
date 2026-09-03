// Shared validation for the two entry points that write
// user.standardAttributes.picture: the avatar dialog on the user detail screen
// (ProfilePictureDialog) and the full-page editor still routed at
// users/:userID/edit-picture (EditPictureScreen). Keep them in one place so the
// two cannot drift apart on accepted type or size.

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
