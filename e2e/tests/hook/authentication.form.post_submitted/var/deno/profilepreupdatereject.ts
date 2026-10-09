export default async function (e: any): Promise<any> {
  return {
    is_allowed: false,
    title: "Profile update rejected",
    reason: "user.profile.pre_update rejected",
  };
}
