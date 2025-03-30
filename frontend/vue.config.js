module.exports = {
    devServer: {
        port: 8081,
        open: true,
        hot: false,
        watchOptions: {
            ignored: /node_modules/,
            poll: true
        }
    }
}