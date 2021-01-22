const path = require("path");
const webpack = require("webpack");

module.exports = {
  mode: "production",
  target: "browserslist",
  entry: {
    authgear: path.resolve(__dirname, "src/authgear.ts"),
    "password-policy": path.resolve(__dirname, "src/password-policy.ts"),
  },
  output: {
    path: path.resolve(__dirname, "../resources/authgear/static"),
    filename: "[name].js"
  },
  plugins: [
    new webpack.ProgressPlugin(),
  ],
  module: {
    rules: [
      {
        test: /\.(js|ts)$/,
        loader: "babel-loader"
      },
    ]
  },
  resolve: {
    extensions: [".js", ".ts"]
  }
};
