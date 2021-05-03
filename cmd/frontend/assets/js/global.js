const $document = $(document);
const $body = $('body');

$(document).on('mouseup', '[data-link] a, [data-link] .select-checkbox, [data-link-tab] a', function (e) {
    e.stopPropagation();
    return false;
});

$('.stop-prop').on('click', function (e) {
    e.stopPropagation();
});

// Data links
let dataLinkDrag = false;
let dataLinkX = 0;
let dataLinkY = 0;

// On document for elements that are created with JS
$document.on('mousedown', '[data-link]', function (e) {
    dataLinkX = e.screenX;
    dataLinkY = e.screenY;
    dataLinkDrag = false;

    if (e.button === 1) {
        return false; // False to stop middle button click dragging
    }
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

    // New window
    if (e.button === 1 || e.ctrlKey || e.shiftKey || e.metaKey || e.which === 2 || target === '_blank') {
        if (!$(e.target).is('a')) {
            window.open(link, '_blank');
        }
        return true;
    }

    window.location.href = link;
    return true;
});

// Console
if (user.isProd) {

    const consoleText = `
 _____                       ____________ 
|  __ \\                      |  _  \\ ___ \\
| |  \\/ __ _ _ __ ___   ___  | | | | |_/ /
| | __ / _\` | '_ \` _ \\ / _ \\ | | | | ___ \\
| |_\\ \\ (_| | | | | | |  __/ | |/ /| |_/ /
 \\____/\\__,_|_| |_| |_|\\___| |___/ \\____/ 
`;

    console.log(consoleText);
}

// Auto dropdowns
const $dropdowns = $('.navbar .dropdown');

$dropdowns.on('mouseenter', function (e) {
    $(this).addClass('show').find('.dropdown-menu').addClass('show');
});
$dropdowns.on('mouseleave', function (e) {
    $(this).removeClass('show').find('.dropdown-menu').removeClass('show');
});
$dropdowns.on('click', function (e) {
    e.stopPropagation();
});

// Tooptips
$body.tooltip({
    selector: '[data-toggle="tooltip"]',
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

    $(function (e) {

        // Choose tab from URL
        const hash = window.location.hash;
        if (hash) {

            let fullHash = '';
            hash.split(/[,\-]/).forEach(function (hash) {
                fullHash = (fullHash === '') ? hash : fullHash + '-' + hash;
                $('.nav-link[href="' + fullHash + '"]').tab('show');
            });
        }

        // Set URL from tab
        $('a[data-toggle="tab"]:not([data-no-hash-change])').on('shown.bs.tab', function (e) {
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
const $top = $('#top');

$(window).on('scroll', function (e) {

    if ($(window).scrollTop() >= 1000) {
        $top.addClass('show');
    } else {
        $top.removeClass('show');
    }
});

$top.on('click', function (e) {
    $('html, body').animate({scrollTop: 0}, 500);
});

// Toasts
if (isIterable(user.toasts)) {
    for (const v of user.toasts) {
        toast(v.success, v.message, v.title, v.timeout, v.link);
    }
}

// Fix URLs
$(function (e) {
    const path = $('#app-page, #package-page, #player-page, #bundle-page, #group-page, #badge-page, #stat-page').attr('data-path');
    if (path && path !== window.location.pathname) {
        history.replaceState(null, null, path + window.location.hash);
    }
});

//
const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

//
function addDataTablesRow(options, data, limit, $table) {

    let $row = $('<tr class="fade-green" />');

    if (typeof options.createdRow === 'function') {
        options.createdRow($row[0], data, null);
    }

    if (isIterable(options.columnDefs)) {
        for (const v of options.columnDefs) {

            let value = data[v];

            if ('visible' in v && v.visible === false) {
                continue;
            }

            if ('render' in v) {
                value = v.render(null, null, data);
            }

            const $td = $('<td />').html(value);

            if ('createdCell' in v) {
                v.createdCell($td, null, data, null, null);
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();

    observeLazyImages($row.find('img[data-lazy]'));
}

// Load AJAX
function loadAjaxOnObserve(map) {

    for (const key in map) {

        if (map.hasOwnProperty(key)) {
            const callback = map[key];
            const element = document.getElementById(key);
            if (element) {
                const f = function (entries, self) {
                    entries.forEach(entry => {
                        if (entry.isIntersecting) {
                            self.unobserve(entry.target);
                            callback();
                        }
                    });
                };
                new IntersectionObserver(f, {rootMargin: '50px 0px 50px 0px', threshold: 0}).observe(element);
            }
        }
    }
}

// (function () {
//
//     // const originalXhr = new window.XMLHttpRequest();
//     const originalXhr = $.ajaxSettings.xhr;
//     $.ajaxSetup({
//         xhr: function () {
//             const xhr = originalXhr();
//             if (xhr) {
//
//                 const $loading = $('#loading');
//
//                 xhr.addEventListener('loadstart', function (e) {
//                     $loading.fadeTo(100, 1);
//                 });
//                 xhr.addEventListener('loadend', function (e) {
//                     $loading.fadeTo(100, 0);
//                 });
//                 xhr.addEventListener('error', function (e) {
//                     logLocal('XHR Error', e)
//                 });
//                 xhr.addEventListener('abort', function (e) {
//                     logLocal('XHR Aborted', e)
//                 });
//             }
//             return xhr;
//         }
//     });
// })();

const cookieName = 'gamedb-session-2';

function setCookieFlag(key, value) {

    let cookieObj = getSessionCookie();

    cookieObj[key] = value;

    return Cookies.set(cookieName, JSON.stringify(cookieObj), {expires: 30, secure: user.isProd});
}

function getSessionCookie(key = null) {

    let cookieText = Cookies.get(cookieName);
    let cookieObj = {};

    if (cookieText) {
        cookieObj = JSON.parse(cookieText);
    }

    if (key) {
        return cookieObj[key];
    } else {
        return cookieObj;
    }
}

$('.jumbotron button.close').on('click', function (e) {
    $(this).closest('.jumbotron').slideUp();
    setCookieFlag($(this).attr('data-id'), true);
});

//
const $darkMode = $('#dark-mode');

$darkMode.on('click', function (e) {

    const $sun = $darkMode.find('.fa-sun');
    const $moon = $darkMode.find('.fa-moon');

    if ($sun.hasClass('d-none')) {

        $sun.removeClass('d-none');
        $moon.addClass('d-none');
        $('body').removeClass('dark');
        setCookieFlag('dark', false);

    } else {

        $sun.addClass('d-none');
        $moon.removeClass('d-none');
        $('body').addClass('dark');
        setCookieFlag('dark', true);
    }

    return false;
});

// Set default dark mode
let darkMode = getSessionCookie('dark');
if (darkMode === undefined && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    $darkMode.trigger('click');
}

// Hide patreon banner
$(document).on('click', '#patreon-message i, #patreon-message svg', function (e) {
    setCookieFlag('patreon-message', true);
    $(this).parent().slideUp(300, function () {
        $('#patreon-message').remove();
    });
    return false;
});

//
function getOS() {

    let os = 'windows';
    if (navigator.appVersion.indexOf('Mac') !== -1) os = 'macos';
    if (navigator.appVersion.indexOf('Linux') !== -1) os = 'linux';

    return os;
}

// Tab links
$('[data-link-tab]').on('mouseup', function () {
    const tab = $(this).attr('data-link-tab');
    $('a.nav-link[href="#' + tab + '"]').tab('show');
    return false;
});

function loadFriends(appID, addSelf, callback) {

    $.ajax({
        type: 'GET',
        url: '/games/' + appID + '/friends.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            const $select = $('select#friend');

            if (data === null) {
                data = [];
            }

            // Sort alphabetically
            data.sort(function (a, b) {
                return a.v.toLowerCase().localeCompare(b.v.toLowerCase());
            });

            $select.empty();
            $select.append('<option value="">Choose Friend</option>');

            if (addSelf && user.playerID && user.playerName) {
                $select.append('<option value="' + user.playerID + '">' + user.playerName + '</option>');
            }

            for (const friend of data) {
                $select.append('<option value="' + friend.k + '">' + friend.v + '</option>');
            }

            const $chosen = $select.chosen({
                disable_search_threshold: 5,
                allow_single_deselect: false,
                max_selected_options: 1,
            });

            $chosen.change(function (e) {
                callback($chosen);
            });
        },
    });
}

window.addEventListener('message', (event) => {
    try {
        let message = JSON.parse(event.data);
        if (message.msg_type === 'resize-me') {

            let shouldCollapseAd = false;

            for (let index in message.key_value) {
                if (message.key_value.hasOwnProperty(index)) {

                    let key = message.key_value[index].key;
                    let value = message.key_value[index].value;

                    if (key === 'r_nh' && value === '0') {
                        shouldCollapseAd = true;
                    }
                }
            }

            if (shouldCollapseAd) {
                $('#flashes-ad').hide();
                logLocal('Ads collapsed');
            }
        }
    } catch (e) {
    }
});
