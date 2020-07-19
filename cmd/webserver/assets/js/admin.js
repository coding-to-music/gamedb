if ($('#admin-tasks-page').length > 0) {

    $('#actions tbody tr').on('click', function () {
        if (confirm('Are you sure?')) {
            $.ajax({
                type: 'get',
                url: $(this).attr('data-action'),
                // success: function (data, textStatus, jqXHR) {
                //     toast(true, 'Triggered');
                // },
                error: function (jqXHR, textStatus, errorThrown) {
                    toast(false, errorThrown);
                },
            });
        }
        return false;
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
                $row.removeClass('table-warning');
                $row.removeClass('table-danger');
                $row.find('.prev').livestamp();
                $row.find('.next').livestamp(new Date(data.Data.time * 1000));
                toast(true, taskID + ' finished', '', 0);
            }
        }
    });
}

if ($('#admin-queues-page').length > 0) {

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
            error: function (jqXHR, textStatus, errorThrown) {
                toast(false, errorThrown);
            },
        });
    });
}

if ($('#admin-users-page').length > 0) {

    const options = {
        "order": [[0, 'desc']],
        "columnDefs": [
            // Date
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return row[0];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Verified
            {
                'targets': 1,
                'render': function (data, type, row) {
                    if (row[2]) {
                        return '<i class="fas fa-check"></i>';
                    } else {
                        return '<i class="fas fa-times"></i>';
                    }
                },
                'orderable': false,
            },
            // Email
            {
                'targets': 2,
                'render': function (data, type, row) {
                    return row[1];
                },
                'orderable': false,
            },
            // Profile
            {
                'targets': 3,
                'render': function (data, type, row) {
                    if (row[3] > 0) {
                        return '<a href="/players/' + row[3] + '">' + row[3] + '</a>';
                    }
                    return '';
                },
                'orderable': false,
            },
            // level
            {
                'targets': 4,
                'render': function (data, type, row) {
                    if (row[4] > 1) {
                        return row[4];
                    }
                    return '';
                },
                'orderSequence': ['desc'],
            },
        ]
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({
        tableOptions: options,
    });
}

if ($('#admin-patreon-page').length > 0) {

    const options = {
        "order": [[0, 'desc']],
        "columnDefs": [
            // Date
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return row[0];
                },
                'orderable': false,
            },
            // Event
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return row[1];
                },
                'orderable': false,
            },
            // User
            {
                'targets': 2,
                'render': function (data, type, row) {
                    return row[2];
                },
                'orderable': false,
            },
        ]
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({
        tableOptions: options,
    });
}
