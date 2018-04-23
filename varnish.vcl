
# varnishd -n /usr/local/var/varnish -f /usr/local/etc/varnish/default.vcl -a 127.0.0.1:8090 -F

vcl 4.0;

import std;

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

    # Send Surrogate-Capability headers to announce ESI support to backend
    set req.http.Surrogate-Capability = "key=ESI/1.0";

    # Strip hash, server doesn't need it.
    if (req.url ~ "\#") {
        set req.url = regsub(req.url, "\#.*$", "");
    }

    # Strip a trailing ? if it exists
    if (req.url ~ "\?$") {
        set req.url = regsub(req.url, "\?$", "");
    }

    # remove double // in urls, /foo and /foo/ are the same url
    # set req.url = regsuball( req.url, "//", "/"      );
    # set req.url = regsub( req.url, "/([?])?$", "\1"  );

    # Normalize the query arguments
    set req.url = std.querysort(req.url);

    # Remove all cookies
    unset req.http.cookie;

    return (hash); // pass pipe synth purge
}

# Happens after we have read the response headers from the backend.
#
# Here you clean the response headers, removing silly Set-Cookie headers
# and other mistakes your backend does.
sub vcl_backend_response {

    // Do ESI processing
    if (bereq.url == "/esi/header") {
        set beresp.ttl = 0s;
    }else{
       set beresp.do_esi = true;
       set beresp.ttl = 24 h;
    }

    // Disable cache
    if (beresp.http.cache-control ~ "(no-cache|private)" || beresp.http.pragma ~ "no-cache") {
        set beresp.ttl = 0s;
    }

    // Remove cookies
    unset beresp.http.set-cookie;

    # Set Expires header based on max-age in Cache-Control
    //set beresp.http.Expires = "" + (now + beresp.ttl);

    # Allow stale content, in case the backend goes down.
    # make Varnish keep all objects for 24 hours beyond their TTL
    set beresp.grace = 24h;

    # Don't cache 50x responses
    if (beresp.status == 500 || beresp.status == 502 || beresp.status == 503 || beresp.status == 504) {
        return (abandon);
    }

    return (deliver); // deliver, abandon, retry
}

# Happens when we have all the pieces we need, and are about to send the
# response to the client.
#
# You can do accounting or modifying the final object here.
sub vcl_deliver {

    # Add debug header to see if it's a HIT/MISS and the number of hits, disable when not needed
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }

    return (deliver); // restart, synth
}
