if ($('#bundles-page').length > 0) {

    const options = $.extend(true, {}, dtDefaultOptions, {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="/assets/img/no-app-image-square.jpg" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
            },
            // Discount
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4] + '%'
                }
            },
            // Apps
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                }
            },
            // Packages
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                }
            },
            // Updated At
            {
                "targets": 4,
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "render": function (data, type, row) {
                    return '<span data-livestamp="' + row[3] + '"></span>';
                }
            }
        ]
    });

    const $table = $('table.table-datatable2');
    const dt = $table.DataTable(options);

    websocketListener('bundles', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
