if ($('#groups-trending-page').length > 0) {

    $('form').on('submit', function (e) {

        $table.DataTable().draw();
        return false;
    });

    $('#type, #errors').on('change', function (e) {

        $table.DataTable().draw();
        return false;
    });

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            data.search = {};
            data.search.search = $('#search').val();
            data.search.type = $('#type').val();
            data.search.errors = $('#errors').val();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
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
                    return '<img data-src="/assets/img/no-app-image-square.jpg" src="' + row[3] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderable": false,
            },
            // Trend Value
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[10].toLocaleString();
                },
                "orderable": false,
            },
            // Trend Chart
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '';
                },
                "orderable": false,
            },
            // Link
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<a href="' + row[8] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                },
                "orderable": false,
            },
        ]
    }));
}
