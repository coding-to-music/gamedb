if ($('#player-page').length > 0) {

    $('#games table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', '/games/' + data[0]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img').attr('data-app-id', rowData[0]);
                }
            },
            // Price
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '$' + row[5];
                }
            },
            // Time
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            },
            // Price/Time
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '$' + row[6];
                }
            }
        ]
    }));
}
