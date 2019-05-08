const $document = $(document);
const $body = $("body");

// Data links
let dataLinkDrag = false;
let dataLinkX = 0;
let dataLinkY = 0;

// On document for elements that are created with JS
$document.on('mousedown', '[data-link]', function (e) {
    dataLinkX = e.screenX;
    dataLinkY = e.screenY;
    dataLinkDrag = false;
});

$document.on('mousemove', '[data-link]', function handler(e) {
    if (!dataLinkDrag && (Math.abs(dataLinkX - e.screenX) > 5 || Math.abs(dataLinkY - e.screenY) > 5)) {
        dataLinkDrag = true;
    }
});

$(document).on('mouseup', '[data-link]', function (e) {

    const link = $(this).attr('data-link');
    const target = $(this).attr('data-target');

    if (!link) {
        return true;
    }

    if (dataLinkDrag) {
        return true;
    }

    // Right click
    if (e.which === 3) {
        return true;
    }

    // Middle click
    if (e.ctrlKey || e.shiftKey || e.metaKey || e.which === 2 || target === '_blank') {

        if (!$(e.target).is("a")) {
            window.open(link, '_blank');
        }
        return true;
    }

    window.location.href = link;
    return true;
});

$('.stop-prop').on('click', function (e) {
    e.stopPropagation();
});

// Auto dropdowns
$('.navbar .dropdown').hover(
    function () {
        $(this).addClass("show").find('.dropdown-menu').addClass("show");
    }, function () {
        $(this).removeClass("show").find('.dropdown-menu').removeClass("show");
    }
).click(function (e) {
    e.stopPropagation();
});

// Tooptips
$body.tooltip({
    selector: '[data-toggle="tooltip"]'
});

// JSON fields
function isJson(str) {
    try {
        JSON.parse(str);
    } catch (e) {
        return false;
    }
    return true;
}

$('.json').each(function (i, value) {

    const json = $(this).text();

    if (isJson(json)) {
        const jsonObj = JSON.parse(json);
        $(this).text(JSON.stringify(jsonObj, null, '\t'));
    }
});

// Tabs
(function ($, window) {
    'use strict';

    $(document).ready(function () {

        // Choose tab from URL
        const hash = window.location.hash;
        if (hash) {

            let fullHash = '';
            hash.split(/[,\-]/).map(function (hash) {

                fullHash = (fullHash === '') ? hash : fullHash + '-' + hash;

                $('.nav-link[href="' + fullHash + '"]').tab('show');
            });
        }

        // Set URL from tab
        $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
            const hash = $(e.target).attr('href');
            if (history.pushState) {
                history.pushState(null, null, hash);
            } else {
                location.hash = hash;
            }
        });
    });

})(jQuery, window);


// Scroll to top link
const $top = $("#top");

$(window).on('scroll', function (e) {

    if ($(window).scrollTop() >= 1000) {
        $top.addClass("show");
    } else {
        $top.removeClass("show");
    }
});

$top.click(function (e) {
    $('html, body').animate({scrollTop: 0}, 500);
});

highLightOwnedGames();

// Websocket helper
function websocketListener(page, onMessage) {

    if (window.WebSocket === undefined) {

        console.log('Your browser does not support websockets');

    } else {

        const socket = new WebSocket((location.protocol === 'https:' ? "wss://gamedb.online" : "ws://" + location.host) + "/websocket/" + page);
        const $badge = $('#live-badge');
        let open = false;

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
            console.log('Websocket opened');
            open = true;
        };

        socket.onclose = function (e) {
            if (open) {
                $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped'); // onerror will trigger too
                console.log('Websocket closed');
            }
        };

        socket.onerror = function (e) {
            if (open) {
                $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped');
            }
        };

        socket.onmessage = onMessage;

        // Click to close websocket manually
        // $badge.on('click', function (e) {
        //     if ($(this).hasClass('cursor-pointer')) {
        //         socket.close(1000);
        //         $badge.addClass('badge-danger').removeClass('badge-secondary badge-success cursor-pointer');
        //         toast(false, 'Live functionality has stopped');
        //     }
        // });
    }
}

// Ads
if (user.showAds) {

    window.CHITIKA = {
        'units': [
            {"calltype": "async[2]", "publisher": "jleagle", "width": 160, "height": 600, "sid": "gamedb-right"},
            {"calltype": "async[2]", "publisher": "jleagle", "width": 160, "height": 600, "sid": "gamedb-left"},
            {"calltype": "async[2]", "publisher": "jleagle", "width": 728, "height": 90, "sid": "gamedb-top-big"},
            {"calltype": "async[2]", "publisher": "jleagle", "width": 320, "height": 50, "sid": "gamedb-top-small"}
        ]
    };

    $('div.container').eq(1)
        .prepend('<div class="ad-right d-none d-xl-block" id="chitikaAdBlock-0"></div>')
        .prepend('<div class="ad-left d-none d-xl-block" id="chitikaAdBlock-1"></div>');
    $('#ad-top')
        .prepend('<div class="ad-top-big d-none d-lg-block d-xl-none" id="chitikaAdBlock-2"></div>')
        .prepend('<div class="ad-top-small d-block d-lg-none" id="chitikaAdBlock-3"></div>');
}

// Toasts
if (isIterable(user.toasts)) {
    for (const v of user.toasts) {
        toast(v.success, v.message, v.title, v.timeout, v.link);
    }
}

// Fix URLs
$(document).ready(function () {
    const path = $('#app-page, #package-page, #player-page, #bundle-page').attr('data-path');
    if (path !== '' && path !== window.location.pathname) {
        history.replaceState(null, null, path);
    }
});

// Broken images
$(document).ready(function () {

    $('img').one('error', function () {

        const url = $(this).attr('data-src');
        if (url) {
            this.src = url;
        }
    });

    $('img[src=""][data-src]').each(function (i, value) {
        this.src = $(this).attr('data-src');
    });
});
