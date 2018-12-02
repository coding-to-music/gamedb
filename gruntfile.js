module.exports = function (grunt) {

    const sass = require('node-sass');

    // Project configuration.
    grunt.initConfig({
        pkg: grunt.file.readJSON('package.json'),
        sass: {
            options: {
                implementation: sass,
                sourceMap: false
            },
            dist: {
                files: {
                    'assets/css/sass/index.css': 'assets/sass/index.scss'
                }
            }
        },
        cssmin: {
            options: {
                sourceMap: false,
                roundingPrecision: -1
            },
            target: {
                files: {
                    'assets/compiled.min.css': ['assets/concatenate.css']
                }
            }
        },
        concat: {
            js: {
                src: [
                    'assets/js/third-party/moment.js', // Put above livestamp
                    'assets/js/third-party/highcharts.min.js', // Put above heatmap
                    'assets/js/third-party/*.js',
                    'assets/js/tabs.js',
                    'assets/js/tables.js',
                    'assets/js/*.js'
                ],
                dest: 'assets/compiled.min.js'
            },
            css: {
                src: [
                    'assets/css/third-party/*.css',
                    'assets/css/sass/*.css',
                    'assets/css/*.css',
                ],
                dest: 'assets/concatenate.css'
            }
        },
        cachebreaker: {
            dev: {
                options: {
                    match: ['assets/compiled.min.css', 'assets/compiled.min.js'],
                },
                files: {
                    src: ['templates/_header.gohtml', 'templates/_footer.gohtml']
                }
            }
        },
        watch: {
            sass: {
                files: ['assets/sass/**/*.scss'],
                tasks: ['sass', 'concat:css', 'cssmin', 'cacheBust', 'clean', 'notify:done']
            },
            js: {
                files: ['assets/js/*.js'],
                tasks: ['concat:js', 'cacheBust', 'clean', 'notify:done']
            }
        },
        clean: [
            'assets/css/sass/',
            'assets/concatenate.css'
        ],
        notify: {
            done: {
                options: {
                    message: 'Done @ ' + new Date().getHours() + ":" + new Date().getMinutes() + ":" + new Date().getSeconds() + '!'
                }
            }
        }
    });

    // Load the plugin that provides the tasks
    grunt.loadNpmTasks('grunt-contrib-concat');
    grunt.loadNpmTasks('grunt-contrib-watch');
    grunt.loadNpmTasks('grunt-contrib-cssmin');
    grunt.loadNpmTasks('grunt-contrib-cssmin');
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-cache-breaker');
    grunt.loadNpmTasks('grunt-notify');
    grunt.loadNpmTasks('grunt-sass');

    // For notify
    grunt.task.run('notify_hooks');

    // Default tasks.
    grunt.registerTask('default', [
        // CSS
        'sass',
        'concat:css',
        'cssmin',

        // JS
        'concat:js',

        //
        'cacheBust',
        'clean',
        'notify:done',
        'watch'
    ]);
};
