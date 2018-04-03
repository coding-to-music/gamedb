if ($('#xp-page').length > 0) {

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
