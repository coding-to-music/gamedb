if ($('#free-games-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[7]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                }
            },
            // Score
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3] + '%';
                }
            },
            // Type
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false,
                "searchable": false
            },
            // Platform
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('platforms platforms-align')
                },
                "orderable": false,
                "searchable": false
            },
            // Install link
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<a href="' + row[6] + '">Install</a>';
                },
                "orderable": false,
                "searchable": false
            }
        ]
    }));

}
