if ($('#packages-page').length > 0) {

    const options = $.extend(true, {}, dtDefaultOptions, {
        "order": [[5, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[8]);
        },
        "columnDefs": [
            // Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[1];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            },
            // Billing Type
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[2];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // License Type
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[3];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Status
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Apps
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                }
            },
            // Updated Time
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[7] + '" data-livestamp="' + row[6] + '">' + row[7] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            }
        ]
    });

    const $table = $('table.table-datatable2');

    const dt = $table.DataTable(options);

    websocketListener('packages', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);
            console.log(data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    })
}
