import ainouCodeStyle from '@ainou/code-style'

export default [
  ...ainouCodeStyle,
  {
    files: ['**/*.{ts,tsx}'],
    rules: {
      // loader/action 中 throw Response 是 React Router 的惯用写法
      '@typescript-eslint/only-throw-error': [
        'error',
        {
          allow: [
            { from: 'lib', name: 'Response' },
            { from: 'file', name: 'CmsErrorResponse' },
          ],
        },
      ],
    },
  },
  {
    ignores: ['build/', 'public/', '.react-router/'],
  },
]
