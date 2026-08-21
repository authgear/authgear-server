export default async function (e: any): Promise<any> {
  const u = e.payload.user;
  if (!u?.id) {
    return {
      is_allowed: false,
      title: "user not resolved",
      reason: "the oidc.jwt.pre_create payload did not carry a resolved user",
    };
  }
  return {
    is_allowed: true,
  };
}
