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
    $('html, body').animate({scrollTop: 0}, 'slow');
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
