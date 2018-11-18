if ($('#upcoming-page').length > 0) {

    // Setup datatable
    $('table.table').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[3, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('data-app-id', rowData[0]);
                },
                "orderable": false
            },
            // Type
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false
            },
            // Price
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5];
                },
                "orderable": false
            },
            // Release Date
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<span data-livestamp="' + row[7] + '"></span>';
                },
                "orderable": false
            }
        ]
    }));
}
