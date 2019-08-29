if ($('#groups-page').length > 0) {

    const $groupsTable = $('table.table');

    $('form').on('submit', function (e) {

        $groupsTable.DataTable().draw();
        return false;
    });

    $('#type, #errors').on('change', function (e) {

        $groupsTable.DataTable().draw();
        return false;
    });

    const options = {
        "ajax": function (data, callback, settings) {

            data.search = {};
            data.search.search = $('#search').val();
            data.search.type = $('#type').val();
            data.search.errors = $('#errors').val();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-group-id64', data[0]);
            $(row).attr('data-group-id', data[11]);
            $(row).attr('data-link', data[2]);
            if (data[7] === 'game' && !$('#type').val()) {
                $(row).addClass('table-primary');
            }
            if (data[9]) {
                $(row).addClass('table-danger');
            }
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-src="/assets/img/no-app-image-square.jpg" data-lazy="' + row[3] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
            },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<a href="' + row[8] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                },
                "orderable": false,
            },
        ]
    };

    $groupsTable.gdbTable({tableOptions: options});
}
