if ($('#xp-page').length > 0) {

    // Scroll to number
    function scroll() {

        if (typeof scrollTo === 'string') {
            window.scroll({
                top: $(scrollTo).offset().top - 100,
                left: 0,
                behavior: 'smooth'
            });

            $('tr').removeClass('table-success');
            $(scrollTo).addClass('table-success');
        }
    }

    $('#xp-page table tr td').click(function () {

        var level = $(this).parent().attr('data-level');

        if (history.pushState) {
            history.pushState('data', '', '/experience/' + level);
        }

        scrollTo = 'tr[data-level=' + level + ']';
        scroll();
    });

    scroll();

    // Calculator
    function levelToXP(level) {
        for (var current = 0, total = 0; current <= level; current++) {
            total += Math.ceil(current / 10) * 100;
        }

        return total;
    }

    function update() {

        var answer = $('#answer');
        answer.val('Loading..');

        var from = $('#from').val();
        if (from < 1) {
            from = 1;
        }

        var to = $('#to').val();
        if (to < 1) {
            to = 1;
        }

        answer.val(levelToXP(to) - levelToXP(from));
    }

    $('#from, #to').change(update);

    $('#calculate').click(function (e) {
        e.preventDefault();
        update();
        return false;
    });

    update();
}