(function ($, window, document, undefined) {
    'use strict';

    // Choose tab from URL
    var hashes = window.location.hash;
    if (hashes) {
        console.log(hashes.split(','));
        hashes.split(',').map(function (hash) {
            $('.nav-link[href="' + hash + '"]').tab('show');
        });
    }

    // Set URL from tab
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
        var hash = $(e.target).attr('href');
        if (history.pushState) {
            history.pushState(null, null, hash);
        } else {
            location.hash = hash;
        }
    });

})(jQuery, window, document);
