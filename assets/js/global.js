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
var $top = $("#top");

function showTopLink() {

    if ($(window).scrollTop() >= 1000) {
        $top.addClass("show");
    } else {
        $top.removeClass("show");
    }
}

$(window).on('scroll', showTopLink);

showTopLink();

$top.click(function () {
    $('html, body').animate({scrollTop: 0}, 500);
});

// Highlight owned games
var games = localStorage.getItem('games');
if (games != null) {
    games = JSON.parse(games);
    if (games != null) {
        $('[data-app-id]').each(function () {
            var id = $(this).attr('data-app-id');
            if (games.indexOf(parseInt(id)) !== -1) {
                $(this).addClass('font-weight-bold')
            }
        });
    }
}

// Browser notification
function browserNotification(message) {

    console.log(message);

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

        var socket = new WebSocket(((location.protocol === 'https:') ? "wss://" : "ws://") + location.host + "/websocket/" + page);
        var $badge = $('#live-badge');

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
