if ($('#admin-page').length > 0) {

    const $actions = $('#actions a');

    $actions.on('click', function () {
        const text = $(this).find('p').text();
        return confirm(text + '?');
    });

    $actions.hover(
        function () {
            $(this).addClass('list-group-item-danger')
        },
        function () {
            $(this).removeClass('list-group-item-danger')
        }
    );

    const queuesForm = $('form#queues');
    queuesForm.on("submit", function (e) {
        e.preventDefault();
        $.ajax({
            type: 'post',
            url: queuesForm.attr('action'),
            data: $(this).serialize()
        });
    });
}
