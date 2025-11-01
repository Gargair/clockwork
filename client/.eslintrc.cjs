/* eslint-env node */
module.exports = {
	env: {
		browser: true,
		es2022: true,
		node: true,
	},
	ignorePatterns: ["dist/", "node_modules/"],
	parser: "@typescript-eslint/parser",
	parserOptions: {
		ecmaVersion: "latest",
		sourceType: "module",
		ecmaFeatures: { jsx: true },
		project: null,
	},
	plugins: ["@typescript-eslint", "react", "react-hooks", "prettier"],
	extends: [
		"eslint:recommended",
		"plugin:@typescript-eslint/recommended",
		"plugin:react/recommended",
		"plugin:react-hooks/recommended",
		"plugin:prettier/recommended",
	],
	settings: {
		react: {
			version: "detect",
		},
	},
	rules: {
		"prettier/prettier": ["warn"],
		"react/react-in-jsx-scope": "off",
		"@typescript-eslint/explicit-function-return-type": ["warn", { allowExpressions: true }],
		"@typescript-eslint/explicit-module-boundary-types": ["warn"],
		"@typescript-eslint/no-explicit-any": ["warn", { ignoreRestArgs: false }],
		"@typescript-eslint/no-unused-vars": [
			"warn",
			{ argsIgnorePattern: "^_", varsIgnorePattern: "^_" }
		],
	},
};


