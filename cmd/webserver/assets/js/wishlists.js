if ($('#wishlists-page').length > 0) {

    const appsOptions = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / App Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Average Position
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["asc", "desc"],
            },
            // Followers
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6];
                },
                "orderSequence": ["desc", "asc"],
            },
            // Price
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[10];
                },
                "orderSequence": ["desc", "asc"],
            },
            // Release Date
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return row[9];
                },
                "orderSequence": ["asc", "desc"],
            },
            // Link
            {
                "targets": 6,
                "render": function (data, type, row) {
                    if (row[7]) {
                        return '<a href="' + row[7] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    $('#apps table.table').gdbTable({
        tableOptions: appsOptions,
        searchFields: [
            $('#search')
        ]
    });

    //
    const tagsOptions = {
        "pageLength": 1000,
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Tag
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<i class="fas fa-tag"></i> ' + row[1];
                },
            },
            // Count
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
        ]
    };

    $('#tags table.table').gdbTable({tableOptions: tagsOptions});
}
