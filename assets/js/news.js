if ($('#news-page').length > 0) {

    // Data tables
    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[3, 'desc']],
        "columnDefs": [
            // Game
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[1] + '" data-livestamp="' + row[0] + '">' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Title
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<i class="fas ' + row[7] + '"></i> ' + row[2];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Author
            {
                "targets": 2,
                "render": function (data, type, row) {

                    if (row[3] === row[6]) {
                        return '<span class="font-weight-bold" data-toggle="tooltip" data-placement="left" title="Your current IP">' + row[3] + '</span>';
                    }
                    return row[3];
                },
                "orderable": false
            },
            // Date
            {
                "targets": 3,
                "render": function (data, type, row) {
                    // return row[4];
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '">' + row[5] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    //$(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            }
        ]
    }));
}
