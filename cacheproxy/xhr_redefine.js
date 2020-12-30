(function() {
    // Redefine XHR to send request to our reverse proxy:
    // TODO
    var proxied = window.XMLHttpRequest.prototype.open;
    window.XMLHttpRequest.prototype.open = function() {
        console.log( arguments[1] );
        return proxied.apply(this, [].slice.call(arguments));
    };
})();
