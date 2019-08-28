if ($('#bundles-page').length > 0) {

    const options = {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);

            // if (data[7]) {
            //     $(row).addClass('table-success');
            // }
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let tagName = row[1];
                    if (row[7]) {
                        tagName = tagName + ' <span class="badge badge-success">Lowest</span>';
                    }

                    return '<div class="icon-name"><div class="icon"><img src="/assets/img/no-app-image-square.jpg" alt="' + row[1] + '"></div><div class="name">' + tagName + '</div></div>'
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
    };


    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    websocketListener('bundles', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
