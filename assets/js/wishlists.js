const $wishlistsPage = $('#wishlists-page');

if ($wishlistsPage.length > 0) {

    $('#apps table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / App
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[4] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
        ]
    }));

    $('#tags table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "pageLength": 1000,
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Tag
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<i class="fas fa-tag"></i> ' + row[1];
                },
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
        ]
    }));
}
