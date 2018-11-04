// Links
$(document).on('mouseup', '[data-link]', function (evnt) {

    var link = $(this).attr('data-link');

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


// Header Country
$('#header-flag').attr('src', '/assets/img/flags/' + user.country.toLowerCase() + '.png').attr('alt', user.country);

// Browser notification
function browserNotification(message) {

    Push.create("Game DB", {
        body: message,
        icon: '/assets/img/sa-bg-32x32.png',
        timeout: 5000,
        vibrate: [100]
    });
}

// Websocket helper
function websocketListener(page, onMessage) {

    if (window.WebSocket === undefined) {

        browserNotification(message);

    } else {

        const socket = new WebSocket(((location.protocol === 'https:') ? "wss://" : "ws://") + location.host + "/websocket/" + page);
        const $badge = $('#live-badge');

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
        };

        socket.onclose = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
            browserNotification('Live functionality has stopped');
        };

        socket.onerror = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
            browserNotification('Live functionality has stopped');
        };

        socket.onmessage = onMessage;
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
