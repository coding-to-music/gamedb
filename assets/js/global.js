// Table row links
$("[data-link]").click(function () {
    var link = $(this).attr('data-link');
    if (link) {
        window.location.href = $(this).attr('data-link');
    }
});

// Clear search on escape
$('input#search').on('keyup', function (evt) {
    var code = evt.charCode || evt.keyCode;
    if (code === 27) {
        $(this).val('');
    }
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
