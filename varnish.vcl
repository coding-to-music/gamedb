
# varnishd -n /usr/local/var/varnish -f /usr/local/etc/varnish/default.vcl -a 127.0.0.1:8090 -F

vcl 4.0;

backend steam {
    .host = "127.0.0.1";
    .port = "8085";
}

# Happens before we check if we have this in cache already.
#
# Typically you clean up the request here, removing cookies you don't need,
# rewriting the request, etc.
sub vcl_recv {

    set req.backend_hint = steam;


}

# Happens after we have read the response headers from the backend.
#
# Here you clean the response headers, removing silly Set-Cookie headers
# and other mistakes your backend does.
sub vcl_backend_response {

}

# Happens when we have all the pieces we need, and are about to send the
# response to the client.
#
# You can do accounting or modifying the final object here.
sub vcl_deliver {

}
