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

    e.stopPropagation();

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

$(document).on('mouseup', '[data-link] a', function (e) {
    e.stopPropagation();
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

//
$('.json').each(function (i, value) {

    const json = $(this).text();

    if (isJson(json)) {
        const jsonObj = JSON.parse(json);
        $(this).text(JSON.stringify(jsonObj, null, '  '));
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

// Toasts
if (isIterable(user.toasts)) {
    for (const v of user.toasts) {
        toast(v.success, v.message, v.title, v.timeout, v.link);
    }
}

// Fix URLs
$(document).ready(function () {
    const path = $('#app-page, #package-page, #player-page, #bundle-page, #group-page').attr('data-path');
    if (path && path !== window.location.pathname) {
        history.replaceState(null, null, path + window.location.hash);
    }
});
