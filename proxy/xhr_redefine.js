(function(proxy_addr) {
    // Redefine XHR to redirect requests to our reverse proxy:
    var proxied = window.XMLHttpRequest.prototype.open;
    window.XMLHttpRequest.prototype.open = function() {
        const stream_regex = /(https?:\/\/)?stream\.navaak\.com\/aes\/_definst_\/.*?\/.*?\.smil\/media_.*?_.*?\.aac/
        let args = [].slice.call(arguments);
        let url = String(args[1]);
        if (stream_regex.test(url)) {
            console.log('Redirected');
            url = url.replace(/(https?:\/\/)?stream\.navaak\.com/, 'http://stream.' + proxy_addr);
        }
        args[1] = url;
        console.log(args);
        return proxied.apply(this, args);
    };
})($PROXY_ADDR$);
