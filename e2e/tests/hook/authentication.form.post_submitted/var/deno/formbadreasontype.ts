export default async function (e: any): Promise<any> {
  return {
    is_allowed: false,
    reasons: [
      {
        type: "unknown",
        field: { user_profile: { pointer: "/given_name" } },
        message: "Unknown type",
      },
    ],
  };
}
