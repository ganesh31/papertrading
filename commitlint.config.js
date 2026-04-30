// commitlint.config.js (at root)
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "scope-enum": [
      2,
      "always",
      [
        // js apps
        "web",
        "api-gateway",
        // go services
        "auth-svc",
        "worker-svc",
        // shared
        "ui",
        "config",
        "deps",
        "ci",
      ],
    ],
  },
};
