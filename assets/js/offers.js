if ($('#offers-page').length > 0) {

    const options = {
        "order": [[0, 'asc']],
        "pageLength": 100,
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[3]);
            $(row).attr('data-app-id', data[0]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Price
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderSequence": ["desc"],
            },
            // Discount
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Rating
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Date
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[7] + '" data-livestamp="' + row[7] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
        ]
    };

    $('table.table').gdbTable({tableOptions: options});
}
