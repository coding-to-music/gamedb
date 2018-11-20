if ($('#admin-page').length > 0) {

    const $actions = $('#actions a');

    $actions.on('click', function () {
        return confirm('Are you sure?');
    });

    const queuesForm = $('form#queues');
    queuesForm.on("submit", function (e) {
        e.preventDefault();
        $.ajax({
            type: 'post',
            url: queuesForm.attr('action'),
            data: $(this).serialize()
        });
    });

    websocketListener('admin', function (e) {

        const data = $.parseJSON(e.data);
        toast(true, data.Data.message, '', 0);
    })
}
