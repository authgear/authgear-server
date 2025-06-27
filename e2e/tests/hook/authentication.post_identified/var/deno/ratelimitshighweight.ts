export default async function (e: any): Promise<any> {
  return {
    is_allowed: true,
    rate_limits: {
      "authentication.general": {
        weight: 100,
      },
    },
  };
}
