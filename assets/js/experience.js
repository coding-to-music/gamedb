if ($('#experience-page').length > 0) {

    const $from = $('#from');
    const $to = $('#to');

    if (user.isLoggedIn) {
        $('.lead span').html('You are level <a href="/experience/' + user.userLevel + '" data-level="' + user.userLevel + '">' + user.userLevel + '</a>.');

        $from.val(user.userLevel);
        $to.val(user.userLevel + 10)
    } else {
        $from.val(10);
        $to.val(20)
    }

    // Scroll to number
    function scroll() {

        if (typeof scrollTo === 'string') {

            const top = $(scrollTo).offset().top - 100;
            $('html, body').animate({scrollTop: top}, 500);

            $('tr').removeClass('table-success');
            $(scrollTo).addClass('table-success');
        }
    }

    $("#experience-page").on("click", "[data-level]", function () {

        const level = $(this).attr('data-level');

        if (history.pushState) {
            history.pushState('data', '', '/experience/' + level);
        }

        scrollTo = 'tr[data-level=' + level + ']';
        scroll();

        return false;
    });

    // Calculator
    function levelToXP(level) {

        let total = 0;

        for (let current = 0; current <= level; current++) {
            total += Math.ceil(current / 10) * 100;
        }

        return total;
    }

    function update() {

        const answer = $('#answer');
        answer.val('Loading..');

        let from = $('#from').val();
        if (from < 1) {
            from = 1;
        }

        let to = $('#to').val();
        if (to < 1) {
            to = 1;
        }

        answer.val((levelToXP(to) - levelToXP(from)).toLocaleString());
    }

    $('#from, #to').change(update);

    $('#calculate').click(function (e) {
        e.preventDefault();
        update();
        return false;
    });

    $(document).ready(scroll);
    $(document).ready(update);
}
