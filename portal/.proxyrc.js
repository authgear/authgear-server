module.exports = function (app) {
  app.use(function (_req, res, next) {
    res.setHeader("Cross-Origin-Embedder-Policy", "unsafe-none");
    next();
  });
};
