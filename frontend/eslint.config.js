import js from "@eslint/js"
import globals from "globals"
import reactHooks from "eslint-plugin-react-hooks"
import reactRefresh from "eslint-plugin-react-refresh"
import tseslint from "typescript-eslint"
import importPlugin from "eslint-plugin-import";

export default tseslint.config(
  {ignores: ["dist"]},
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      "import": importPlugin,
    },
    settings: {
      "import/resolver": {
        "node": {
          "extensions": [".js", ".jsx", ".ts", ".tsx"],
        },
      },
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      "react-refresh/only-export-components": [
        "warn",
        {allowConstantExport: true},
      ],
      "semi": ["error", "always"],
      "quotes": ["error", "double"],
      "indent": ["error", 2],
      "eol-last": ["error", "always"],
      "comma-dangle": ["error", "always-multiline"],
      "no-trailing-spaces": "error",
      "object-curly-spacing": ["error", "always"],
      "array-bracket-spacing": ["error", "never"],
      "@typescript-eslint/no-unused-vars": ["warn", {argsIgnorePattern: "^_"}],
      "@typescript-eslint/explicit-function-return-type": "off",
      "import/extensions": [
        "error",
        "ignorePackages",
        {
          "js": "never",
          "jsx": "never",
          "ts": "never",
          "tsx": "never",
        },
      ],
      "import/order": [
        "error",
        {
          "groups": [
            "builtin",
            "external",
            "internal",
            ["parent", "sibling", "index"],
            "object",
            "type"
          ],
          "pathGroups": [
            {
              pattern: "react",
              group: "external",
              position: "before"
            }
          ],
          "pathGroupsExcludedImportTypes": ["react"],
          "newlines-between": "always",
          "alphabetize": {
            order: "asc",
            caseInsensitive: true
          }
        }
      ]
    },
  },
)
