if ($('#price-changes-page').length > 0) {

    const options = $.extend(true, {}, dtDefaultOptions, {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[5]);

            let x = Math.min(data[9], 100); // Get a range of -100 to 100
            x += 100; // Get a range of 0 to 200
            x = x / 2; // Get a range of 0 to 100
            $(row).addClass('col-grad-' + Math.round(x));

            // if (data[12] > 0) {
            // $(row).addClass('table-danger');
            // } else if (data[12] < 0) {
            // $(row).addClass('table-success');
            // }
        },
        "columnDefs": [
            // App/Package Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[4] + '" class="rounded square" alt="' + row[3] + '"><span>' + row[3] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img').attr('data-app-id', 0)
                },
                "orderable": false
            },
            // Before
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[6];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // After
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[7];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Change
            {
                "targets": 3,
                "render": function (data, type, row) {

                    const small = '<small>' + row[9] + '%</small>';

                    if (row[9] === 0) {
                        return row[8];
                    }

                    return row[8] + ' ' + small;
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Time
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[11] + '">' + row[10] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            }
        ]
    });

    const $table = $('table.table-datatable2');
    const dt = $table.DataTable(options);

    websocketListener('prices', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
