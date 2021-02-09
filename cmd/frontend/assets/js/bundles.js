if ($('#bundles-page').length > 0) {

    const options = {
        "order": [[5, 'desc']],
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
                "orderSequence": ["asc", "desc"],
            },
            // Price
            {
                "targets": 2,
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "render": function (data, type, row) {
                    if (user.prodCC in row[9]) {
                        return row[9][user.prodCC];
                    }
                    return '-';
                },
                "orderable": false,
            },
            // Apps
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Packages
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Updated At
            {
                "targets": 5,
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "render": function (data, type, row) {
                    return '<span data-livestamp="' + row[3] + '"></span>';
                }
            },
            // Link
            {
                "targets": 6,
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
