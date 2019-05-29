const $badgePage = $('#badge-page');

if ($badgePage.length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[5]);
        },
        "columnDefs": [
            // Ranks
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[0];
                },
            },
            // Icon / Player
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
            },
            // Level
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
            // Time
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            },
        ]
    }));
}
