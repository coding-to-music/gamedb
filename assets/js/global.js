// Tablw row links
$("[data-link]").click(function () {
    var link = $(this).attr('data-link');
    if (link) {
        window.location.href = $(this).attr('data-link');
    }
});

// Clear search on escape
function clearField(evt, input) {
    var code = evt.charCode || evt.keyCode;
    if (code === 27) {
        input.value = '';
    }
}

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
