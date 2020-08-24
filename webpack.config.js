// Core
const path = require('path');

// Plugins
const UglifyJSPlugin = require('uglifyjs-webpack-plugin');
// const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const OptimizeCssAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const ConcatPlugin = require('webpack-concat-plugin');
const WebpackBuildNotifierPlugin = require('webpack-build-notifier');

// Config
module.exports = {
    watch: true,
    watchOptions: {
        aggregateTimeout: 100,
        ignored: /node_modules/,
    },
    mode: 'production',
    entry: [
        path.resolve(__dirname, 'cmd/frontend/assets/sass/index.scss'),
    ],
    output: {
        path: path.resolve(__dirname, 'cmd/frontend/assets/dist'),
        publicPath: '/assets/dist/',
    },
    // devtool: "source-map",
    optimization: {
        minimizer: [
            new OptimizeCssAssetsPlugin({
                cssProcessorOptions: {
                    cssProcessorOptions: {
                        map: {
                            inline: false,
                            annotation: false
                        }
                    }
                }
            }),
            new UglifyJSPlugin({
                sourceMap: false,
                parallel: true,
                cache: true,
            }),
        ],
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                use: [],
            },
            {
                test: /\.(s*)css$/,
                use: [
                    {
                        loader: MiniCssExtractPlugin.loader,
                        options: {
                            sourceMap: false,
                        },
                    },
                    {
                        loader: "css-loader",
                        options: {
                            sourceMap: false,
                        },
                    },
                    {
                        loader: "sass-loader",
                        options: {
                            // name: 'css/[name].blocks.css',
                            sourceMap: false,
                            // minimize: true,
                            implementation: require("node-sass"),
                            sassOptions: {
                                includePaths: [
                                    path.resolve(__dirname, 'cmd/frontend/assets/sass/*'),
                                ],
                            }
                        },
                    }
                ],
            }
        ]
    },
    plugins: [
        new MiniCssExtractPlugin(
            {
                filename: "main.css",
            }
        ),
        new OptimizeCssAssetsPlugin(
            {
                assetNameRegExp: /\.(s*)css$/,
                cssProcessor: require('cssnano'),
                cssProcessorOptions: {},
                cssProcessorPluginOptions: {
                    preset: ['default', {discardComments: {removeAll: true}}],
                },
                canPrint: true
            }
        ),
        new ConcatPlugin({
            uglify: true,
            sourceMap: false,
            outputPath: './',
            fileName: 'main.js',
            filesToConcat: [
                path.resolve(__dirname, 'cmd/frontend/assets/js/third-party/*.js'),
                path.resolve(__dirname, 'cmd/frontend/assets/js/helpers/*.js'),
                path.resolve(__dirname, 'cmd/frontend/assets/js/plugins/*.js'),
                path.resolve(__dirname, 'cmd/frontend/assets/js/global.js'),
                path.resolve(__dirname, 'cmd/frontend/assets/js/product.js'),
                path.resolve(__dirname, 'cmd/frontend/assets/js/*.js'),
            ],
            attributes: {
                async: true
            }
        }),
        // new UglifyJSPlugin(
        //     {
        //         sourceMap: true
        //     }
        // ),
        // new HtmlWebpackPlugin(
        //     {
        //         filename: path.resolve(__dirname, 'cmd/frontend/templates/_webpack_header.gohtml'),
        //         template: path.resolve(__dirname, 'cmd/frontend/templates/_webpack_header.ejs'),
        //         hash: true,
        //         inject: false,
        //     }
        // ),
        // new HtmlWebpackPlugin(
        //     {
        //         filename: path.resolve(__dirname, 'cmd/frontend/templates/_webpack_footer.gohtml'),
        //         template: path.resolve(__dirname, 'cmd/frontend/templates/_webpack_footer.ejs'),
        //         hash: true,
        //         inject: false,
        //     }
        // ),
        new WebpackBuildNotifierPlugin(
            {
                // title: "My Project Webpack Build",
                // logo: path.resolve("./img/favicon.png"),
                // suppressSuccess: true
            }
        ),
    ],
};
