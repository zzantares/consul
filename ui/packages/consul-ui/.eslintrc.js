module.exports = {
  root: true,
  parser: 'babel-eslint',
  parserOptions: {
    ecmaVersion: 2018,
    sourceType: 'module',
    ecmaFeatures: {
      legacyDecorators: true,
    },
  },
  plugins: ['ember', 'qunit'],
  extends: ['eslint:recommended', 'plugin:ember/recommended', 'plugin:prettier/recommended'],
  env: {
    browser: true,
  },
  rules: {
    'no-unused-vars': ['warn', { args: 'none' }],
    'ember/no-test-import-export': ['warn'],
    'ember/require-tagless-components': ['warn'],
    'ember/no-empty-glimmer-component-classes': ['warn'],
    'ember/no-private-routing-service': ['warn'],
    'ember/no-computed-properties-in-native-classes': ['warn'],
    'ember/no-actions-hash': ['warn'],
    'ember/no-controller-access-in-routes': ['warn'],
    'ember/no-component-lifecycle-hooks': ['warn'],
    'ember/no-classic-classes': ['warn'],
    'ember/no-classic-components': ['warn'],
    'ember/no-legacy-test-waiters': ['warn'],
    'ember/no-get': ['warn'],
    'ember/no-mixins': ['warn'],
    'ember/no-new-mixins': ['warn'],
    'ember/classic-decorator-no-classic-methods': ['warn'],
    'ember/classic-decorator-hooks': ['warn'],
    'ember/no-test-this-render': ['warn'],
    'ember/no-test-module-for': ['warn'],
    'ember/no-jquery': 'warn',
    'ember/no-global-jquery': 'warn',
    'qunit/no-only': 'error',
  },
  overrides: [
    // node files
    {
      files: [
        '.eslintrc.js',
        '.dev.eslintrc.js',
        '.docfy-config.js',
        '.prettierrc.js',
        '.template-lintrc.js',
        'ember-cli-build.js',
        'testem.js',
        'blueprints/*/index.js',
        'config/**/*.js',
        'lib/*/index.js',
        'server/**/*.js',
      ],
      parserOptions: {
        sourceType: 'script',
      },
      env: {
        browser: false,
        node: true,
      },
      plugins: ['node'],
      rules: Object.assign({}, require('eslint-plugin-node').configs.recommended.rules, {
        // add your custom rules and overrides for node files here

        // this can be removed once the following is fixed
        // https://github.com/mysticatea/eslint-plugin-node/issues/77
        'node/no-unpublished-require': 'off',
      }),
    },
  ],
};
