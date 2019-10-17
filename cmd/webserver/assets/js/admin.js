if ($('#admin-page').length > 0) {

    $('#actions tr').on('click', function () {
        if (confirm('Are you sure?')) {
            $.ajax({
                type: 'get',
                url: $(this).attr('data-action'),
                // success: function (data, textStatus, jqXHR) {
                //     toast(true, 'Triggered');
                // },
                error: function (jqXHR, textStatus, errorThrown) {
                    toast(true, errorThrown);
                },
            });
        }
        return false;
    });

    const queuesForm = $('form#queues');
    queuesForm.on("submit", function (e) {
        e.preventDefault();
        $.ajax({
            type: 'post',
            url: queuesForm.attr('action'),
            data: $(this).serialize(),
            success: function (data, textStatus, jqXHR) {
                toast(true, 'Queued');
                queuesForm.trigger("reset");
            },
        });
    });

    websocketListener('admin', function (e) {

        const data = JSON.parse(e.data);

        const taskID = data.Data.task_id;
        const action = data.Data.action;

        if (taskID && action) {

            const $row = $('tr[data-id="' + taskID + '"]');
            if (action === 'started') {
                $row.addClass('table-warning');
                $row.removeClass('table-danger');
                // toast(true, taskID + ' started', '', 0);
            } else if (action === 'finished') {
                $row.removeClass('table-warning', 'table-danger');
                $row.find('.prev').livestamp();
                $row.find('.next').livestamp(new Date(data.Data.time * 1000));
                // toast(true, taskID + ' finished', '', 0);
            }
        }
    });
}
