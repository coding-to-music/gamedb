if ($('#new-releases-page').length > 0) {

    const options = {
        "order": [[3, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / App Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img alt="" data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
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
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderSequence": ["desc", "asc"],
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
                "orderSequence": ["desc", "asc"],
            },
            // Players
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[8].toLocaleString();
                },
                "orderSequence": ["desc", "asc"],
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
                "orderSequence": ["desc", "asc"],
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
                "orderSequence": ["desc", "asc"],
            },
        ]
    };

    // Init table
    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: [
            $('#search'),
            $('#days'),
        ]
    });
}
