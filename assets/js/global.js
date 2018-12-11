// Links
$(document).on('mouseup', '[data-link]', function (evnt) {

    const link = $(this).attr('data-link');

    if (evnt.which === 3) {
        return true;
    }

    if (evnt.ctrlKey || evnt.shiftKey || evnt.metaKey || evnt.which === 2) {
        window.open(link, '_blank');
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
$("body").tooltip({
    selector: '[data-toggle="tooltip"]'
});

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

// Highlight owned games
function highLightOwnedGames() {
    let games = localStorage.getItem('games');
    if (games != null) {
        games = JSON.parse(games);
        if (games != null) {
            $('[data-app-id]').each(function () {
                const id = $(this).attr('data-app-id');
                if (games.indexOf(parseInt(id)) !== -1) {
                    $(this).addClass('font-weight-bold')
                }
            });
        }
    }
}

highLightOwnedGames();

// Websocket helper
function websocketListener(page, onMessage) {

    if (window.WebSocket === undefined) {

        toast(false, 'Your browser does not support websockets');

    } else {

        const socket = new WebSocket(((location.protocol === 'https:') ? "wss://" : "ws://") + location.host + "/websocket/" + page);
        const $badge = $('#live-badge');

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
        };

        socket.onclose = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
            toast(false, 'Live functionality has stopped');
        };

        socket.onerror = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
            toast(false, 'Live functionality has stopped');
        };

        socket.onmessage = onMessage;

        $badge.on('click', function (e) {
            if ($(this).hasClass('cursor-pointer')) {
                socket.close(1000);
                $badge.addClass('badge-danger').removeClass('badge-secondary badge-success cursor-pointer');
                toast(true, 'Live functionality stopped');
            }
        });
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

function toast(success = true, body, title = '', timeout = 8, link = '') {

    const redirect = function () {
        if (link === 'refresh') {
            link = window.location.href;
        }
        window.location.replace(link);

    };

    const options = {
        timeOut: Number(timeout) * 1000,
        onclick: link ? redirect : null,

        newestOnTop: true,
        preventDuplicates: false,
        extendedTimeOut: 0, // Keep alive on hover
    };

    if (success) {
        toastr.success(body, title, options);
    } else {
        toastr.error(body, title, options);
    }

}

function isIterable(obj) {
    // checks for null and undefined
    if (obj == null) {
        return false;
    }
    return typeof obj[Symbol.iterator] === 'function';
}

// Flag
const flag = $('<img src="/assets/img/flags/' + user.country.toLowerCase() + '.png" alt="' + user.country + '">');
if (user.isLoggedIn) {
    $('#header-flag').html(flag);
} else {
    $('#header-flag').html('<a href="/login">' + flag.prop('outerHTML') + '</a>');
}

// Admin link
if (user.isAdmin) {
    $('#header-admin').html('<a class="nav-link" href="/admin">Admin</a>');
}

// User link
const $headerUser = $('#header-user');
const $headerSettings = $('#header-settings');

if (user.isLoggedIn) {
    $headerUser.html('<a class="nav-link" href="/players/' + user.userID + '">' + user.userName + '</a>');

    $headerSettings.prepend('<div class="dropdown-divider"></div>');
    $headerSettings.prepend('<a class="dropdown-item" href="/logout"><i class="fas fa-sign-out-alt"></i> Logout</a>');
    $headerSettings.prepend('<a class="dropdown-item" href="/settings"><i class="fas fa-cog"></i> Settings</a>');
} else {
    $headerUser.html('<a class="nav-link" href="/login">Login</a>');
}

// Flashes
if (isIterable(user.flashesGood)) {
    let $flashesGood = $('#flashes-good');
    for (const v of user.flashesGood) {
        $flashesGood.append('<p>' + v + '</p>');
        $flashesGood.removeClass('d-none');
    }
}

if (isIterable(user.flashesBad)) {
    let $flashesBad = $('#flashes-bad');
    for (const v of user.flashesBad) {
        $flashesBad.append('<p>' + v + '</p>');
        $flashesBad.removeClass('d-none');
    }
}
