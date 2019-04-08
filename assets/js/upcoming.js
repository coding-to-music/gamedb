if ($('#upcoming-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        // "order": [[3, 'asc']],
        // "pageLength": 2000,
        "paging": false,
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "drawCallback": function (settings) {
            const api = this.api();
            const rows = api.rows({page: 'current'}).nodes();

            let last = null;
            api.rows().every(function (rowIdx, tableLoop, rowLoop) {
                let group = this.data()[6];
                if (last !== group) {
                    $(rows).eq(rowIdx).before(
                        '<tr class="table-success"><td colspan="3">' + group + '</td></tr>'
                    );
                    last = group;
                }
            });
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
                },
                "orderable": false,
            },
            // App Type
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Price
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Release Date
            // {
            //     "targets": 3,
            //     "render": function (data, type, row) {
            //         return row[6];
            //     },
            //     "createdCell": function (td, cellData, rowData, row, col) {
            //         $(td).attr('nowrap', 'nowrap');
            //     },
            //     "orderable": false,
            // },
        ]
    }));
}
