if ($('#bundles-page').length > 0) {

    const options = {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / Bundle Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let tagName = row[1];
                    if (row[7]) {
                        tagName = tagName + ' <span class="badge badge-success">Lowest</span>';
                    }

                    return '<a href="' + row[2] + '" class="icon-name"><div class="icon"><img src="/assets/img/no-app-image-square.jpg" alt="' + row[1] + '"></div><div class="name">' + tagName + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Discount
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4] + '%'
                },
                "orderSequence": ["asc"],
            },
            // Apps
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Packages
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderSequence": ["desc"],
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
            },
            // Link
            {
                "targets": 5,
                "render": function (data, type, row) {
                    if (row[8]) {
                        return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };


    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    websocketListener('bundles', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
