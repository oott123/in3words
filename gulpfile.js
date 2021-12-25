const gulp = require('gulp')
const sass = require('gulp-sass')(require('sass'))
const postcss = require('gulp-postcss')

function fonts() {
  return gulp
    .src('node_modules/@fontsource/*/files/*.{ttf,woff,woff2,eot,svg}')
    .pipe(gulp.dest('public/_i3w_remix/gulp/fonts'))
}

function styles() {
  return gulp
    .src('styles/index.scss')
    .pipe(
      sass
        .sync({
          includePaths: ['node_modules'],
        })
        .on('error', sass.logError),
    )
    .pipe(
      postcss({
        plugins: [require('tailwindcss'), require('autoprefixer')],
      }),
    )
    .pipe(gulp.dest('app/styles'))
}

function watchStyles() {
  return gulp.watch('styles/**/*.scss', { ignoreInitial: false }, styles)
}

const watch = gulp.parallel([watchStyles])
const defaultTask = gulp.parallel([fonts, watch])

exports.watch = watch
exports.styles = styles
exports.fonts = fonts
exports.build = gulp.parallel([fonts, styles])
exports.default = defaultTask
