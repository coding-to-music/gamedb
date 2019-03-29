if ($('#new-releases-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "pageLength": 50,
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('data-app-id', rowData[0]);
                },
            },
            // Price
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
            },
            // Score
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[7] + '%';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
            },
            // Players
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[8].toLocaleString();
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
            },
            // Release Date
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[6];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
            },
            // Chart
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return '<div data-app-id="' + row[0] + '"><i class="fas fa-spinner fa-spin"></i></div>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('chart');
                },
                "orderable": false,
            },
        ]
    }));
}
