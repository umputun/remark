module.exports = {
  parser: 'babel-eslint',
  extends: [
    'eslint:recommended',
    'plugin:jsx-a11y/recommended',
    'plugin:prettier/recommended',
    'plugin:@typescript-eslint/recommended',
  ],
  plugins: ['react', 'jsx-a11y', 'prettier'],
  overrides: [
    {
      files: ['*.ts', '*.tsx'],
      plugins: ['@typescript-eslint'],
      parser: '@typescript-eslint/parser',
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: './',
      },
      rules: {
        // disabling because typescipt uses it's own lint (see next rule)
        'no-unused-vars': 0,
        // allow Rust-like var starting with _underscore
        '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: /^_/ }],
        // disabling because it's bad practice to mark acceesibility in react classes
        '@typescript-eslint/explicit-member-accessibility': 0,
        // doesn't work in real world
        '@typescript-eslint/no-non-null-assertion': 0,
        // actually it better to disable inly in tests, but i'm tired to to do this now, feel free to change
        '@typescript-eslint/no-explicit-any': 0,
        // disabling because store actions use WATCH_ME_IM_SPECIAL case
        '@typescript-eslint/class-name-casing': 0,
        // actually it better to disable inly in tests, but i'm tired to to do this now, feel free to change
        '@typescript-eslint/no-object-literal-type-assertion': 0,
        // disabling because server output uses snake case response
        '@typescript-eslint/camelcase': 0,
        // disabling because it's standard behaviour that function is hoisted to top
        '@typescript-eslint/no-use-before-define': 0,
        // maybe good but I have just tired to type return types everywhere, especially with complex generic return types
        '@typescript-eslint/explicit-function-return-type': 0,
      },
    },
  ],
  env: {
    browser: true,
    node: true,
    es6: true,
    jest: true,
  },
  parserOptions: {
    ecmaVersion: 6,
    sourceType: 'module',
    ecmaFeatures: {
      modules: true,
      jsx: true,
    },
  },
  globals: {
    remark_config: true,
  },
  rules: {
    '@typescript-eslint/indent': 0,
    'react/jsx-uses-react': 2,
    'react/jsx-uses-vars': 2,
    'no-cond-assign': 1,
    'no-empty': 0,
    'no-console': 1,
    camelcase: 0,
    'comma-style': 2,
    'max-nested-callbacks': [2, 3],
    'no-eval': 2,
    'no-implied-eval': 2,
    'no-new-func': 2,
    'guard-for-in': 0,
    eqeqeq: 0,
    'no-else-return': 2,
    'no-redeclare': 2,
    'no-dupe-keys': 2,
    radix: 2,
    strict: [2, 'never'],
    'no-shadow': 0,
    'callback-return': [1, ['callback', 'cb', 'next', 'done']],
    'no-delete-var': 2,
    'no-undef-init': 2,
    'no-shadow-restricted-names': 2,
    'handle-callback-err': 0,
    'no-lonely-if': 0,
    'constructor-super': 2,
    'no-this-before-super': 2,
    'no-dupe-class-members': 2,
    'no-const-assign': 2,
    'prefer-spread': 2,
    'prefer-const': 2,
    'no-useless-concat': 2,
    'no-var': 2,
    'object-shorthand': 2,
    'prefer-arrow-callback': 2,
    'prettier/prettier': 2,
    '@typescript-eslint/no-var-requires': 0,
  },
};
