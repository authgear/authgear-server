export default {
  preset: "ts-jest",
  testEnvironment: "jsdom",
  transform: {
    "^.+\\.tsx?$": "ts-jest",
  },
  testPathIgnorePatterns: ["/node_modules/", "<rootDir>/comments-api/"],
  transformIgnorePatterns: ["<rootDir>/node_modules/"],
  moduleNameMapper: {
    "\\.(svg|png|jpg|jpeg|gif)$": "<rootDir>/src/__mocks__/fileMock.js",
  },
};
