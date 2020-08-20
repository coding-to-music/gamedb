const $xpPage = $('#experience-page');

if ($xpPage.length > 0) {

    // Scroll to number
    function scroll() {

        if (typeof scrollTo === 'string') {

            const top = $(scrollTo).offset().top - 100;
            $('html, body').animate({scrollTop: top}, 500);

            $('tr').removeClass('table-success');
            $(scrollTo).addClass('table-success');
        }
    }

    $xpPage.on("click", "tr[data-level]", function (e) {

        const level = $(this).attr('data-level');

        if (history.pushState) {
            history.pushState('data', '', '/experience/' + level);
        }

        scrollTo = 'tr[data-level=' + level + ']';
        scroll();
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

    $('#from, #to').on('change', update);

    $('#calculate').on('click', function (e) {
        e.preventDefault();
        update();
    });

    $(function (e) {
        scroll();
        update();
    });
}
